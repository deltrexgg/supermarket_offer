// Smoke test for a running stack: public API, login, roles, tenant isolation and visibility.
//
//   node scripts/smoke.mjs
//
// Settings come from the environment, with the .env file in the project folder as fallback:
//   API_URL (default http://localhost:8080), SUPABASE_URL, SUPABASE_ANON_KEY,
//   ADMIN_EMAIL + ADMIN_PASSWORD (optional: without them the admin checks are skipped).
// Needs the demo data (SEED=true): it signs in as noor@offers.test and gulf@offers.test.
//
// With Docker instead of Node on your computer:
//   docker run --rm -v "${PWD}:/p" -w /p -e API_URL=http://host.docker.internal:8080 node:22-alpine node scripts/smoke.mjs

import { readFileSync } from "node:fs";

function loadDotEnv() {
  try {
    for (const line of readFileSync(new URL("../.env", import.meta.url), "utf8").split(/\r?\n/)) {
      const m = line.match(/^\s*([A-Z0-9_]+)\s*=\s*(.*?)\s*$/);
      if (m && process.env[m[1]] === undefined) process.env[m[1]] = m[2].replace(/^["']|["']$/g, "");
    }
  } catch {
    /* no .env file: use the environment only */
  }
}
loadDotEnv();

const API = (process.env.API_URL ?? "http://localhost:8080").replace(/\/$/, "") + "/api/v1";
const HEALTH = API.replace("/api/v1", "/health");
const SUPABASE_URL = (process.env.SUPABASE_URL ?? "").replace(/\/$/, "");
const ANON = process.env.SUPABASE_ANON_KEY ?? "";
const DEMO_PASSWORD = "Offers@12345";

let passed = 0;
let failed = 0;
let skipped = 0;

function check(name, ok, detail = "") {
  if (ok) {
    passed++;
    console.log(`  ok    ${name}`);
  } else {
    failed++;
    console.log(`  FAIL  ${name}${detail ? "  (" + detail + ")" : ""}`);
  }
}
function skip(name, why) {
  skipped++;
  console.log(`  skip  ${name}  (${why})`);
}

async function call(path, { method = "GET", token, body, raw = false } = {}) {
  const headers = {};
  if (token) headers.Authorization = `Bearer ${token}`;
  if (body !== undefined) headers["Content-Type"] = "application/json";
  const res = await fetch(raw ? path : API + path, { method, headers, body: body === undefined ? undefined : JSON.stringify(body) });
  const text = await res.text();
  let json = null;
  try {
    json = JSON.parse(text);
  } catch {
    /* not JSON */
  }
  return { status: res.status, json, headers: res.headers };
}

async function signIn(email, password) {
  if (!SUPABASE_URL || !ANON) return null;
  const res = await fetch(`${SUPABASE_URL}/auth/v1/token?grant_type=password`, {
    method: "POST",
    headers: { apikey: ANON, "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) return null;
  return (await res.json()).access_token ?? null;
}

const errorCode = (r) => r.json?.error?.code;

async function publicApi() {
  console.log("\nPublic API");
  const health = await call(HEALTH, { raw: true });
  check("GET /health is 200", health.status === 200, `status ${health.status}`);

  const list = await call("/offers");
  check("GET /offers is 200 with items", list.status === 200 && Array.isArray(list.json?.items), `status ${list.status}`);
  check("every listed offer is active and not expired", (list.json?.items ?? []).every((o) => o.days_left >= 0));
  const again = await call("/offers");
  check("second identical request is served from Redis", again.headers.get("x-cache") === "HIT", `x-cache ${again.headers.get("x-cache")}`);

  for (const path of ["/areas", "/categories", "/supermarkets"]) {
    const r = await call(path);
    check(`GET ${path} is 200`, r.status === 200 && Array.isArray(r.json?.items), `status ${r.status}`);
  }

  const bad = await call("/offers?sort=bogus");
  check("unknown sort is 422 in the standard error format", bad.status === 422 && errorCode(bad) === "validation_failed", `status ${bad.status}`);
  const notNumber = await call("/offers/abc");
  check("GET /offers/abc is 404", notNumber.status === 404 && errorCode(notNumber) === "not_found", `status ${notNumber.status}`);
  const missing = await call("/offers/999999999");
  check("GET /offers/999999999 is 404", missing.status === 404, `status ${missing.status}`);
  const unknown = await call("/nope");
  check("unknown route is a JSON 404", unknown.status === 404 && errorCode(unknown) === "not_found", `status ${unknown.status}`);

  const headers = list.headers;
  check("security headers are set", headers.get("x-content-type-options") === "nosniff" && headers.get("x-frame-options") === "DENY");
  return list.json?.items ?? [];
}

async function auth(tokens) {
  console.log("\nSign-in is required");
  for (const path of ["/portal/offers", "/portal/branches", "/admin/stats", "/auth/me"]) {
    const r = await call(path);
    check(`GET ${path} without a token is 401`, r.status === 401, `status ${r.status}`);
  }
  const junk = await call("/auth/me", { token: "not.a.token" });
  check("a made-up token is 401", junk.status === 401, `status ${junk.status}`);
  const me = await call("/auth/me", { token: tokens.noor });
  check("signed-in responses are not cached (no-store)", (me.headers.get("cache-control") ?? "").includes("no-store"));
}

async function roles(tokens) {
  console.log("\nRoles");
  const own = await call("/portal/offers", { token: tokens.noor });
  check("supermarket can read its portal offers", own.status === 200, `status ${own.status}`);
  const adminAsShop = await call("/admin/stats", { token: tokens.noor });
  check("supermarket gets 403 on /admin/stats", adminAsShop.status === 403, `status ${adminAsShop.status}`);

  if (!tokens.admin) {
    skip("admin role checks", "set ADMIN_EMAIL and ADMIN_PASSWORD");
    return;
  }
  const stats = await call("/admin/stats", { token: tokens.admin });
  check("admin can read /admin/stats", stats.status === 200 && typeof stats.json?.offers_active === "number", `status ${stats.status}`);
  const shopAsAdmin = await call("/portal/offers", { token: tokens.admin });
  check("admin gets 403 on the supermarket portal", shopAsAdmin.status === 403, `status ${shopAsAdmin.status}`);
}

async function tenantIsolation(tokens) {
  console.log("\nTenant isolation");
  const noor = await call("/portal/offers?page_size=50", { token: tokens.noor });
  const gulf = await call("/portal/offers?page_size=50", { token: tokens.gulf });
  const noorIds = new Set((noor.json?.items ?? []).map((o) => o.id));
  const gulfIds = (gulf.json?.items ?? []).map((o) => o.id);
  check("both supermarkets have offers to compare", noorIds.size > 0 && gulfIds.length > 0);
  check("the two offer lists do not overlap", gulfIds.every((id) => !noorIds.has(id)));
  if (gulfIds.length === 0) return;

  const target = gulfIds[0];
  const read = await call(`/portal/offers/${target}`, { token: tokens.noor });
  check("reading another supermarket's offer is 404", read.status === 404, `status ${read.status}`);
  const del = await call(`/portal/offers/${target}`, { method: "DELETE", token: tokens.noor });
  check("deleting another supermarket's offer is 404", del.status === 404, `status ${del.status}`);
  const poster = await call(`/portal/offers/${target}/poster`, { method: "POST", token: tokens.noor, body: { template: "hot_deal", lang: "en" } });
  check("making a poster for another supermarket's offer is 404", poster.status === 404, `status ${poster.status}`);
  const still = await call(`/portal/offers/${target}`, { token: tokens.gulf });
  check("the offer is still there for its owner", still.status === 200, `status ${still.status}`);
}

async function visibility(tokens) {
  console.log("\nPublic visibility follows supermarket status");
  if (!tokens.admin) {
    skip("pending/blocked supermarket visibility", "set ADMIN_EMAIL and ADMIN_PASSWORD");
    return;
  }
  const me = await call("/auth/me", { token: tokens.gulf });
  const shopId = me.json?.supermarket_id;
  if (!shopId) {
    check("gulf login has a supermarket", false);
    return;
  }
  const setStatus = (status) =>
    call(`/admin/supermarkets/${shopId}/status`, { method: "PATCH", token: tokens.admin, body: { status } });

  const before = await call(`/offers?supermarket=${shopId}`);
  check("approved supermarket's offers are public", (before.json?.total ?? 0) > 0, `total ${before.json?.total}`);
  const sample = before.json?.items?.[0]?.id;

  try {
    for (const status of ["blocked", "pending"]) {
      const set = await setStatus(status);
      check(`admin can set the supermarket to ${status}`, set.status === 200, `status ${set.status}`);
      const hidden = await call(`/offers?supermarket=${shopId}`);
      check(`${status} supermarket has no public offers`, hidden.json?.total === 0, `total ${hidden.json?.total}`);
      if (sample) {
        const detail = await call(`/offers/${sample}`);
        check(`${status} supermarket's offer detail is 404`, detail.status === 404, `status ${detail.status}`);
      }
    }
  } finally {
    const restore = await setStatus("approved");
    check("supermarket is approved again", restore.status === 200, `status ${restore.status}`);
  }
  const after = await call(`/offers?supermarket=${shopId}`);
  check("its offers are public again", (after.json?.total ?? 0) > 0, `total ${after.json?.total}`);
}

async function main() {
  console.log(`Smoke test against ${API}`);
  try {
    await publicApi();
  } catch (e) {
    console.log(`\nCannot reach the API: ${e.message}. Is the stack running?`);
    process.exit(2);
  }

  const tokens = {
    noor: await signIn("noor@offers.test", DEMO_PASSWORD),
    gulf: await signIn("gulf@offers.test", DEMO_PASSWORD),
    admin: process.env.ADMIN_EMAIL && process.env.ADMIN_PASSWORD ? await signIn(process.env.ADMIN_EMAIL, process.env.ADMIN_PASSWORD) : null,
  };
  if (!tokens.noor || !tokens.gulf) {
    console.log("\nCould not sign in as noor@ and gulf@offers.test. Set SUPABASE_URL and SUPABASE_ANON_KEY and run with SEED=true data.");
    console.log("The login-dependent checks were not run.");
    summary();
    process.exit(1);
  }
  if (process.env.ADMIN_EMAIL && !tokens.admin) console.log("\nWarning: the admin login in .env did not work; admin checks are skipped.");

  await auth(tokens);
  await roles(tokens);
  await tenantIsolation(tokens);
  await visibility(tokens);
  summary();
  process.exit(failed ? 1 : 0);
}

function summary() {
  console.log(`\n${passed} passed, ${failed} failed, ${skipped} skipped`);
}

main();
