# supermarket_offer

Offer platform for supermarkets: a merchant portal and a public deals website.

| Part | Tech | Port |
|---|---|---|
| `api/` | Go, Gin, GORM | 8080 |
| `web-portal/` | Next.js (supermarket portal) | 3000 |
| `web-admin/` | Next.js (platform owner console) | 3002 |
| `web-public/` | Next.js (public website) | 3001 |
| Database | PostgreSQL on Supabase | hosted |
| Login | Supabase Auth | hosted |
| Images | Supabase Storage (S3 bucket) | hosted |
| redis | Redis 7 (docker-compose) | 6379 |

Design docs are in `docs/` (`SCHEMA.md`, `API.md`, `MODULES.md`).

## Set up Supabase (once)

1. Create a Supabase project.
2. **Database:** Project Settings > Database > Connection string > **Session pooler**. Copy it into `DATABASE_URL` (port 5432, `sslmode=require`). Do not use the transaction pooler (port 6543): the migration runner needs a session-level lock and GORM uses prepared statements.
3. **Auth:** Project Settings > API. Copy the project URL, the public (anon) key and the service-role key. The service-role key is secret and only the API uses it.
4. **Storage:** create a **public** bucket named `offer-media`, then Storage > S3 Connection to get the endpoint, region and access keys.
5. `cp .env.example .env` and fill in the values. Never commit `.env`.

## Run

```bash
docker compose up --build
```

On start the API applies the SQL migrations in `api/migrations/` to your Supabase database (look for `migrate: applied 0001_init.sql` in the log).

With `SEED=true` (the default in `.env.example`) the API also fills an empty database with Dubai demo data (3 supermarkets, 6 branches, 30 offers with dates relative to today) and creates these supermarket logins in your Supabase Auth. Password for all: `Offers@12345`.

The platform owner login is not part of the demo data. Set `ADMIN_EMAIL` and `ADMIN_PASSWORD` in `.env`; the API creates that login on every start if it does not exist (with or without `SEED`).

| Login | Role |
|---|---|
| `noor@offers.test` | Al Noor Fresh Mart |
| `gulf@offers.test` | Gulf Basket |
| `spice@offers.test` | Spice Garden Hypermarket |

Set `SEED=false` afterwards. The seed never touches a database that already has supermarkets.

- API health: http://localhost:8080/health
- Portal: http://localhost:3000 (sign in with a demo login, or register a new supermarket; branches, offers and photo upload)
- Admin: http://localhost:3002 (sign in with `ADMIN_EMAIL` / `ADMIN_PASSWORD`: approve supermarkets, manage areas and categories, review offers)
- Public site: http://localhost:3001 (search, filters by area, category and supermarket, sort, paging, offer detail with branches; responses cached in Redis, see the `X-Cache` header on `/api/v1/offers`). Arabic with right-to-left layout is at http://localhost:3001/ar, with a language switch in the header. `/robots.txt` and `/sitemap.xml` are generated too; the sitemap lists every active offer in both languages and is rebuilt when requested (cached ten minutes).

## HTML to image

The API image bundles Chromium. `chromedp` drives it headless to render HTML as an image:

```bash
curl -X POST http://localhost:8080/api/v1/render \
  -H 'Content-Type: application/json' \
  -d '{"html":"<h1 style=\"font-family:sans-serif\">Fresh Milk 20% off</h1>","width":800,"height":400}' \
  --output flyer.png
```

Options: `width`, `height`, `scale` (max 3), `format` (`png` or `jpeg`), `quality` (jpeg).

## Settings (`.env`)

| Setting | Needed | What it does |
|---|---|---|
| `DATABASE_URL` | yes | Supabase session pooler connection string. Percent-encode special characters in the password. |
| `SUPABASE_URL`, `SUPABASE_ANON_KEY`, `SUPABASE_SERVICE_ROLE_KEY` | yes | Login. The anon key goes to the portal and admin builds; the service-role key stays in the API. |
| `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY` | yes | Supabase Storage through its S3 API. |
| `PUBLIC_FILE_BASE_URL` | yes | Where browsers load photos and posters: `https://REF.supabase.co/storage/v1/object/public/offer-media`. The bucket must be public. |
| `ADMIN_EMAIL`, `ADMIN_PASSWORD` | no | The platform owner login, created on startup. Set both or neither. Password 8 to 72 characters. |
| `SEED` | no | `true` fills an empty database with demo data and creates the demo supermarket logins. |
| `CORS_ORIGINS` | yes | Web app addresses allowed to call the API: ports 3000, 3001 and 3002 locally. |
| `NEXT_PUBLIC_API_URL` | yes | API address the browsers use. Baked into the web builds, so rebuild after changing it. |
| `ANTHROPIC_API_KEY`, `ANTHROPIC_MODEL` | no | Claude writes the Arabic item name and the English and Arabic descriptions when a supermarket clicks "Suggest with Claude" in the portal. Without a key the button explains it is not set up and everything else works. The model defaults to `claude-haiku-5-5`. |
| `SITE_URL` | no | The public address of the shopper website (no trailing slash), used for canonical links, `robots.txt` and `sitemap.xml`. Default `http://localhost:3001`. Baked into the build, so rebuild after changing it. |
| `TRUSTED_PROXIES` | no | Proxy addresses or CIDRs whose `X-Forwarded-For` is believed. Leave empty unless the API sits behind a reverse proxy you control. |
| `RATE_LIMIT_PUBLIC` | no | Public requests per minute per IP, default 600. `0` turns it off. |

## Tests

Unit tests cover the logic that needs no database: Dubai dates, the tenant filter, image type detection, poster HTML, cache keys and the rate limiter. They run inside the API build image, so Go is not needed on your computer:

```bash
docker build --target build -t offers-api-build ./api
docker run --rm offers-api-build go test ./...
```

The smoke test checks a running stack end to end: public endpoints, sign-in required, roles, one supermarket not reaching another's data, and blocked or pending supermarkets staying off the public site. It needs the demo data (`SEED=true`) and `SUPABASE_URL` and `SUPABASE_ANON_KEY` in `.env`. Add `ADMIN_EMAIL` and `ADMIN_PASSWORD` to include the admin checks. It briefly sets Gulf Basket to blocked and pending, and always approves it again at the end.

```bash
node scripts/smoke.mjs
# or, without Node on your computer:
docker run --rm -v "${PWD}:/p" -w /p -e API_URL=http://host.docker.internal:8080 node:22-alpine node scripts/smoke.mjs
```

## Security notes

- Passwords are handled only by Supabase Auth. The API verifies the signed access token and loads the user's role and supermarket from its own `users` table. A valid login with no row there gets nothing.
- The supermarket always comes from that verified row, never from the URL or request body. Another supermarket's record is answered with `404`.
- Uploads are checked by file signature (JPEG, PNG, WebP, GIF, AVIF; max 5 MB) and stored under random names.
- Rate limits per IP, counted in Redis: public API (`RATE_LIMIT_PUBLIC`), 300 per minute for signed-in routes, 120 for `/admin`, 20 for poster creation, 10 for `/render`, 10 per hour for registration. Over the limit gives `429` with `Retry-After`. If Redis is down the limits are skipped.
- Behind a reverse proxy, set `TRUSTED_PROXIES`. Without it the socket address is used as the client IP, so a forged `X-Forwarded-For` cannot dodge the limits. Note that the public website's own server calls the API from one container, so its visitors share one counter. That is why the public limit is high, and the answers are cached in Redis.
- Responses carry `X-Content-Type-Options`, `X-Frame-Options` and `Referrer-Policy`. Signed-in responses are `no-store`.
- Database row-level security is on with no policies, so Supabase's own public data API can read nothing.
- A **blocked** or **pending** supermarket is hidden from the public site, but it can still sign in to its portal. Refusing the portal as well is not built yet.

## Notes

- `api/go.sum` is generated during the Docker build. To pin versions: `docker build --target build -t offers-api-build ./api`, then `docker create --name tmp offers-api-build`, `docker cp tmp:/src/go.sum api/go.sum`, `docker cp tmp:/src/go.mod api/go.mod`, `docker rm tmp`, and commit both files.
- Everything here uses Dubai time (`Asia/Dubai`) for "today", whatever the server or browser time zone is.
- Arabic names, descriptions and area, category and supermarket names: the first two are per offer; area, category and supermarket names stay in English on the Arabic site because the database has no Arabic columns for them.

## Troubleshooting

- `prepared statement "stmtcache_..." already exists` on start: the Supabase pooler reused a connection that still held statements from an earlier run. The API now uses the simple query protocol (no prepared statements), so this should not happen. If you see it, you are running an older image: `docker compose up --build`.
- `lookup ...@aws-0-...: no such host`: the database password in `DATABASE_URL` has special characters. Percent-encode them (`@` is `%40`) or use a password without any.
- The admin or portal login page loops back to sign in and the browser console shows a CORS error: add that app's address (for example `http://localhost:3002`) to `CORS_ORIGINS` in `.env`, then `docker compose up -d api`.
