"use client";

import { useCallback, useEffect, useState } from "react";
import { api, type Page, type Shop, type ShopStatus } from "@/lib/api";

const TABS: [string, string][] = [["", "All"], ["pending", "Pending"], ["approved", "Approved"], ["blocked", "Blocked"]];

export default function SupermarketsPage() {
  const [status, setStatus] = useState("");
  const [q, setQ] = useState("");
  const [page, setPage] = useState(1);
  const [data, setData] = useState<Page<Shop> | null>(null);
  const [error, setError] = useState("");
  const [busyId, setBusyId] = useState<number | null>(null);

  // The dashboard links here with ?status=...; read it once on the client.
  useEffect(() => {
    const s = new URLSearchParams(window.location.search).get("status");
    if (s) setStatus(s);
  }, []);

  const load = useCallback(async () => {
    const qs = new URLSearchParams({ page: String(page) });
    if (status) qs.set("status", status);
    if (q.trim()) qs.set("q", q.trim());
    try {
      setData(await api<Page<Shop>>(`/admin/supermarkets?${qs}`));
      setError("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not load");
    }
  }, [status, q, page]);

  useEffect(() => {
    const t = setTimeout(load, q ? 250 : 0);
    return () => clearTimeout(t);
  }, [load, q]);

  async function setShopStatus(s: Shop, next: ShopStatus) {
    if (next === "blocked" && !window.confirm(`Block ${s.name}? Its offers disappear from the public site.`)) return;
    setBusyId(s.id);
    try {
      await api(`/admin/supermarkets/${s.id}/status`, { method: "PATCH", body: { status: next } });
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not update");
    } finally {
      setBusyId(null);
    }
  }

  const pages = data ? Math.max(1, Math.ceil(data.total / data.page_size)) : 1;

  return (
    <>
      <h1>Supermarkets</h1>
      <div className="row">
        <div className="tabs">
          {TABS.map(([v, label]) => (
            <button key={v} className={status === v ? "active" : ""} onClick={() => { setStatus(v); setPage(1); }}>{label}</button>
          ))}
        </div>
        <input className="search" placeholder="Search by name" value={q} onChange={(e) => { setQ(e.target.value); setPage(1); }} aria-label="Search supermarkets" />
      </div>
      {error && <p className="error">{error}</p>}
      {!data ? <p className="muted">Loading…</p> : data.items.length === 0 ? (
        <p className="muted">No supermarkets match.</p>
      ) : (
        <table>
          <thead>
            <tr><th>Name</th><th>Login</th><th>Registered</th><th>Branches</th><th>Offers</th><th>Status</th><th /></tr>
          </thead>
          <tbody>
            {data.items.map((s) => (
              <tr key={s.id}>
                <td><strong>{s.name}</strong>{s.phone && <div className="muted">{s.phone}</div>}</td>
                <td>{s.email}</td>
                <td>{s.created_at}</td>
                <td>{s.branch_count}</td>
                <td>{s.offer_count}</td>
                <td><span className={`status ${s.status}`}>{s.status}</span></td>
                <td className="actions">
                  {s.status !== "approved" && (
                    <button className="primary" disabled={busyId === s.id} onClick={() => setShopStatus(s, "approved")}>
                      {s.status === "blocked" ? "Unblock" : "Approve"}
                    </button>
                  )}
                  {s.status !== "blocked" && (
                    <button className="danger" disabled={busyId === s.id} onClick={() => setShopStatus(s, "blocked")}>
                      {s.status === "pending" ? "Reject" : "Block"}
                    </button>
                  )}
                </td>
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
