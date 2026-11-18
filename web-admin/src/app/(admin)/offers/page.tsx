"use client";

import { useCallback, useEffect, useState } from "react";
import { api, type AdminOffer, type Page } from "@/lib/api";
import { aed, statusLabel } from "@/lib/format";

const TABS: [string, string][] = [["", "All"], ["active", "Active"], ["upcoming", "Upcoming"], ["expired", "Expired"]];

export default function OffersPage() {
  const [status, setStatus] = useState("");
  const [q, setQ] = useState("");
  const [page, setPage] = useState(1);
  const [data, setData] = useState<Page<AdminOffer> | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    const s = new URLSearchParams(window.location.search).get("status");
    if (s) setStatus(s);
  }, []);

  const load = useCallback(async () => {
    const qs = new URLSearchParams({ page: String(page) });
    if (status) qs.set("status", status);
    if (q.trim()) qs.set("q", q.trim());
    try {
      setData(await api<Page<AdminOffer>>(`/admin/offers?${qs}`));
      setError("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not load");
    }
  }, [status, q, page]);

  useEffect(() => {
    const t = setTimeout(load, q ? 250 : 0);
    return () => clearTimeout(t);
  }, [load, q]);

  async function remove(o: AdminOffer) {
    if (!window.confirm(`Remove "${o.item_name}" from ${o.supermarket_name}? This cannot be undone.`)) return;
    try {
      await api(`/admin/offers/${o.id}`, { method: "DELETE" });
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not remove");
    }
  }

  const pages = data ? Math.max(1, Math.ceil(data.total / data.page_size)) : 1;

  return (
    <>
      <h1>Offers</h1>
      <p className="muted">Offers from every supermarket. Remove only what breaks the rules; shops manage their own offers.</p>
      <div className="row">
        <div className="tabs">
          {TABS.map(([v, label]) => (
            <button key={v} className={status === v ? "active" : ""} onClick={() => { setStatus(v); setPage(1); }}>{label}</button>
          ))}
        </div>
        <input className="search" placeholder="Search item or brand" value={q} onChange={(e) => { setQ(e.target.value); setPage(1); }} aria-label="Search offers" />
      </div>
      {error && <p className="error" role="alert">{error}</p>}
      {!data ? <p className="muted">Loading…</p> : data.items.length === 0 ? (
        <p className="muted">No offers match.</p>
      ) : (
        <table>
          <thead>
            <tr><th>Item</th><th>Supermarket</th><th>Price</th><th>Valid</th><th>Status</th><th /></tr>
          </thead>
          <tbody>
            {data.items.map((o) => (
              <tr key={o.id}>
                <td><strong>{o.item_name}</strong><div className="muted">{[o.brand, o.unit].filter(Boolean).join(" · ")}</div></td>
                <td>{o.supermarket_name}</td>
                <td>{aed(o.offer_price)} <span className="strike">{aed(o.regular_price)}</span> <span className="badge">-{o.discount_pct}%</span></td>
                <td>{o.valid_from} to {o.valid_to}</td>
                <td><span className={`status ${o.status}`}>{statusLabel(o.status, o.days_left)}</span></td>
                <td className="actions"><button className="danger" onClick={() => remove(o)}>Remove</button></td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {pages > 1 && (
        <div className="row">
          <button disabled={page <= 1} onClick={() => setPage(page - 1)}>Previous</button>
          <span className="muted">Page {page} of {pages}</span>
          <button disabled={page >= pages} onClick={() => setPage(page + 1)}>Next</button>
        </div>
      )}
    </>
  );
}
