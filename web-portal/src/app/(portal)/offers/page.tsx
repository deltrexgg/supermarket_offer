"use client";

import { useCallback, useEffect, useState } from "react";
import { api, apiBlob, ApiError, type Branch, type Category, type Offer, type Page } from "@/lib/api";
import { aed, dubaiToday, statusLabel } from "@/lib/format";

const TABS = [
  { key: "", label: "All" },
  { key: "active", label: "Active" },
  { key: "upcoming", label: "Upcoming" },
  { key: "expired", label: "Expired" },
];

type Draft = {
  id?: number;
  item_name: string; brand: string; unit: string; category_id: string;
  regular_price: string; offer_price: string; valid_from: string; valid_to: string;
  branch_ids: number[]; photoUrl: string | null;
};

export default function OffersPage() {
  const [status, setStatus] = useState("");
  const [search, setSearch] = useState("");
  const [term, setTerm] = useState("");
  const [page, setPage] = useState(1);
  const [data, setData] = useState<Page<Offer> | null>(null);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [draft, setDraft] = useState<Draft | null>(null);
  const [photo, setPhoto] = useState<File | null>(null);
  const [fields, setFields] = useState<Record<string, string>>({});
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [posterFor, setPosterFor] = useState<Offer | null>(null);
  const [tpl, setTpl] = useState("hot_deal");
  const [lang, setLang] = useState("en");
  const [posterUrl, setPosterUrl] = useState<string | null>(null);
  const [posterBusy, setPosterBusy] = useState(false);
  const [posterError, setPosterError] = useState("");

  // Wait a moment after typing before searching.
  useEffect(() => {
    const t = setTimeout(() => { setTerm(search); setPage(1); }, 300);
    return () => clearTimeout(t);
  }, [search]);

  const load = useCallback(async () => {
    const qs = new URLSearchParams({ page: String(page), page_size: "10" });
    if (status) qs.set("status", status);
    if (term) qs.set("q", term);
    try {
      setData(await api<Page<Offer>>(`/portal/offers?${qs}`));
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not load offers.");
    }
  }, [status, term, page]);

  useEffect(() => { load(); }, [load]);

  useEffect(() => {
    (async () => {
      try {
        const [b, c] = await Promise.all([
          api<{ items: Branch[] }>("/portal/branches"),
          api<{ items: Category[] }>("/categories", { auth: false }),
        ]);
        setBranches(b.items);
        setCategories(c.items);
      } catch { /* the list shows its own error */ }
    })();
  }, []);

  function openNew() {
    setPosterFor(null);
    const today = dubaiToday();
    setFields({}); setError(""); setPhoto(null);
    setDraft({
      item_name: "", brand: "", unit: "", category_id: "", regular_price: "", offer_price: "",
      valid_from: today, valid_to: today, branch_ids: [], photoUrl: null,
    });
  }

  function openEdit(o: Offer) {
    setPosterFor(null);
    setFields({}); setError(""); setPhoto(null);
    setDraft({
      id: o.id, item_name: o.item_name, brand: o.brand ?? "", unit: o.unit, category_id: String(o.category_id),
      regular_price: o.regular_price.toFixed(2), offer_price: o.offer_price.toFixed(2),
      valid_from: o.valid_from.slice(0, 10), valid_to: o.valid_to.slice(0, 10),
      branch_ids: o.branch_ids, photoUrl: o.photo_url,
    });
  }

  async function save(e: React.FormEvent) {
    e.preventDefault();
    if (!draft) return;
    setBusy(true); setError(""); setFields({});
    const fd = new FormData();
    fd.set("item_name", draft.item_name);
    fd.set("brand", draft.brand);
    fd.set("unit", draft.unit);
    fd.set("category_id", draft.category_id);
    fd.set("regular_price", draft.regular_price);
    fd.set("offer_price", draft.offer_price);
    fd.set("valid_from", draft.valid_from);
    fd.set("valid_to", draft.valid_to);
    draft.branch_ids.forEach((id) => fd.append("branch_ids[]", String(id)));
    if (photo) fd.set("photo", photo);
    try {
      if (draft.id) await api(`/portal/offers/${draft.id}`, { method: "PUT", body: fd });
      else await api("/portal/offers", { method: "POST", body: fd });
      setDraft(null);
      await load();
    } catch (err) {
      if (err instanceof ApiError) {
        setFields(err.fields);
        setError(Object.keys(err.fields).length ? "" : err.message);
      } else setError("Could not save the offer.");
    }
    setBusy(false);
  }

  async function remove(o: Offer) {
    if (!window.confirm(`Delete the offer "${o.item_name}"?`)) return;
    try {
      await api(`/portal/offers/${o.id}`, { method: "DELETE" });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not delete the offer.");
    }
  }

  function openPoster(o: Offer) {
    setDraft(null);
    setPosterFor(o);
    setPosterUrl(o.poster_url);
    setPosterError("");
  }

  async function makePoster() {
    if (!posterFor) return;
    setPosterBusy(true);
    setPosterError("");
    try {
      const r = await api<{ poster_url: string }>(`/portal/offers/${posterFor.id}/poster`, {
        method: "POST", body: { template: tpl, lang },
      });
      setPosterUrl(r.poster_url);
      await load();
    } catch (err) {
      setPosterError(err instanceof Error ? err.message : "Could not create the poster.");
    }
    setPosterBusy(false);
  }

  async function downloadPoster() {
    if (!posterFor) return;
    try {
      const blob = await apiBlob(`/portal/offers/${posterFor.id}/poster`);
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `poster-${posterFor.id}.png`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      setPosterError(err instanceof Error ? err.message : "Could not download the poster.");
    }
  }

  async function printPoster() {
    if (!posterFor) return;
    try {
      const blob = await apiBlob(`/portal/offers/${posterFor.id}/poster/print?template=${tpl}&lang=${lang}`);
      window.open(URL.createObjectURL(new Blob([blob], { type: "text/html" })), "_blank");
    } catch (err) {
      setPosterError(err instanceof Error ? err.message : "Could not open the print view.");
    }
  }

  const toggleBranch = (id: number) => {
    if (!draft) return;
    const has = draft.branch_ids.includes(id);
    setDraft({ ...draft, branch_ids: has ? draft.branch_ids.filter((b) => b !== id) : [...draft.branch_ids, id] });
  };
  const set = (k: keyof Draft) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
    draft && setDraft({ ...draft, [k]: e.target.value });

  const totalPages = data ? Math.max(1, Math.ceil(data.total / data.page_size)) : 1;
  const preview = photo ? URL.createObjectURL(photo) : draft?.photoUrl;

  return (
    <>
      <div className="row">
        <h1>Offers</h1>
        <span className="spacer" />
        <button className="primary" onClick={openNew} disabled={branches.length === 0}>New offer</button>
      </div>
      {branches.length === 0 && <p className="muted">Add a branch first: offers are sold at your branches.</p>}
      {error && !draft && <p className="error" role="alert">{error}</p>}

      {draft && (
        <form className="panel" onSubmit={save}>
          <h2>{draft.id ? "Edit offer" : "New offer"}</h2>
          <div className="grid">
            <label>Item name
              <input value={draft.item_name} onChange={set("item_name")} required />
              {fields.item_name && <span className="error" role="alert">{fields.item_name}</span>}
            </label>
            <label>Brand (optional)
              <input value={draft.brand} onChange={set("brand")} />
              {fields.brand && <span className="error" role="alert">{fields.brand}</span>}
            </label>
            <label>Unit (e.g. 1 L, 5 kg)
              <input value={draft.unit} onChange={set("unit")} required />
              {fields.unit && <span className="error" role="alert">{fields.unit}</span>}
            </label>
            <label>Category
              <select value={draft.category_id} onChange={set("category_id")} required>
                <option value="">Choose a category</option>
                {categories.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
              {fields.category_id && <span className="error" role="alert">{fields.category_id}</span>}
            </label>
            <label>Regular price (AED)
              <input inputMode="decimal" value={draft.regular_price} onChange={set("regular_price")} required />
              {fields.regular_price && <span className="error" role="alert">{fields.regular_price}</span>}
            </label>
            <label>Offer price (AED)
              <input inputMode="decimal" value={draft.offer_price} onChange={set("offer_price")} required />
              {fields.offer_price && <span className="error" role="alert">{fields.offer_price}</span>}
            </label>
            <label>Valid from
              <input type="date" value={draft.valid_from} onChange={set("valid_from")} required />
              {fields.valid_from && <span className="error" role="alert">{fields.valid_from}</span>}
            </label>
            <label>Valid to
              <input type="date" value={draft.valid_to} onChange={set("valid_to")} required />
              {fields.valid_to && <span className="error" role="alert">{fields.valid_to}</span>}
            </label>
          </div>

          <fieldset>
            <legend>Available at</legend>
            {branches.map((b) => (
              <label key={b.id} className="check">
                <input type="checkbox" checked={draft.branch_ids.includes(b.id)} onChange={() => toggleBranch(b.id)} />
                {b.name} <span className="muted">({b.area_name})</span>
              </label>
            ))}
            {fields.branch_ids && <span className="error" role="alert">{fields.branch_ids}</span>}
          </fieldset>

          <label>Photo (JPEG, PNG, WebP, GIF or AVIF, max 5 MB)
            <input type="file" accept="image/jpeg,image/png,image/webp,image/gif,image/avif" onChange={(e) => setPhoto(e.target.files?.[0] ?? null)} />
          </label>
          {/* eslint-disable-next-line @next/next/no-img-element */}
          {preview && <img className="thumb big" src={preview} alt="Offer photo preview" />}

          {error && <p className="error" role="alert">{error}</p>}
          <div className="row">
            <button className="primary" disabled={busy}>{busy ? "Saving…" : "Save offer"}</button>
            <button type="button" onClick={() => setDraft(null)}>Cancel</button>
          </div>
        </form>
      )}

      {posterFor && (
        <div className="panel">
          <h2>Poster for {posterFor.item_name}</h2>
          <div className="row">
            <label>Template
              <select value={tpl} onChange={(e) => setTpl(e.target.value)}>
                <option value="hot_deal">Hot deal</option>
                <option value="weekend_offer">Weekend offer</option>
              </select>
            </label>
            <label>Language
              <select value={lang} onChange={(e) => setLang(e.target.value)}>
                <option value="en">English</option>
                <option value="ar">Arabic</option>
              </select>
            </label>
            <button className="primary" onClick={makePoster} disabled={posterBusy}>
              {posterBusy ? "Creating…" : posterUrl ? "Create again" : "Create poster"}
            </button>
          </div>
          {posterError && <p className="error" role="alert">{posterError}</p>}
          {posterUrl && (
            <>
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img className="poster-preview" src={posterUrl} alt="Offer poster preview" />
              <div className="row">
                <button onClick={downloadPoster}>Download PNG</button>
                <button onClick={printPoster}>Print</button>
              </div>
              <p className="muted">Print uses the template and language selected above.</p>
            </>
          )}
          <div className="row"><button onClick={() => setPosterFor(null)}>Close</button></div>
        </div>
      )}

      <div className="row">
        <div className="tabs">
          {TABS.map((t) => (
            <button key={t.key} className={status === t.key ? "active" : ""} onClick={() => { setStatus(t.key); setPage(1); }}>{t.label}</button>
          ))}
        </div>
        <span className="spacer" />
        <input className="search" placeholder="Search item or brand" value={search} onChange={(e) => setSearch(e.target.value)} />
      </div>

      {data === null ? <p className="muted">Loading…</p> : data.items.length === 0 ? (
        <p className="muted">No offers here yet.</p>
      ) : (
        <table>
          <thead><tr><th /><th>Item</th><th>Price</th><th>Valid</th><th>Status</th><th>Branches</th><th /></tr></thead>
          <tbody>
            {data.items.map((o) => (
              <tr key={o.id}>
                <td>{/* eslint-disable-next-line @next/next/no-img-element */}
                  {o.photo_url ? <img className="thumb" src={o.photo_url} alt="" /> : <span className="thumb empty" />}</td>
                <td><strong>{o.item_name}</strong><br /><span className="muted">{[o.brand, o.unit].filter(Boolean).join(" · ")}</span></td>
                <td>{aed(o.offer_price)} <span className="strike">{aed(o.regular_price)}</span> <span className="badge">-{o.discount_pct}%</span></td>
                <td>{o.valid_from.slice(0, 10)} → {o.valid_to.slice(0, 10)}</td>
                <td><span className={`status ${o.status}`}>{statusLabel(o.status, o.days_left)}</span></td>
                <td>{o.branch_ids.length}</td>
                <td className="actions">
                  <button onClick={() => openPoster(o)}>Poster</button>
                  <button onClick={() => openEdit(o)}>Edit</button>
                  <button className="danger" onClick={() => remove(o)}>Delete</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {data && data.total > data.page_size && (
        <div className="row">
          <button disabled={page <= 1} onClick={() => setPage(page - 1)}>Previous</button>
          <span className="muted">Page {page} of {totalPages}</span>
          <button disabled={page >= totalPages} onClick={() => setPage(page + 1)}>Next</button>
        </div>
      )}
    </>
  );
}
