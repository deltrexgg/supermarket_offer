"use client";

import { useCallback, useEffect, useState } from "react";
import { api, ApiError, type Area, type Branch } from "@/lib/api";

type Draft = { id?: number; name: string; area_id: string; address: string; phone: string };
const empty: Draft = { name: "", area_id: "", address: "", phone: "" };

export default function BranchesPage() {
  const [branches, setBranches] = useState<Branch[] | null>(null);
  const [areas, setAreas] = useState<Area[]>([]);
  const [draft, setDraft] = useState<Draft | null>(null);
  const [fields, setFields] = useState<Record<string, string>>({});
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    try {
      const [b, a] = await Promise.all([
        api<{ items: Branch[] }>("/portal/branches"),
        api<{ items: Area[] }>("/areas", { auth: false }),
      ]);
      setBranches(b.items);
      setAreas(a.items);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not load branches.");
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  function edit(b: Branch) {
    setFields({});
    setError("");
    setDraft({ id: b.id, name: b.name, area_id: String(b.area_id), address: b.address, phone: b.phone ?? "" });
  }

  async function save(e: React.FormEvent) {
    e.preventDefault();
    if (!draft) return;
    setBusy(true);
    setError("");
    setFields({});
    const body = { name: draft.name, area_id: Number(draft.area_id), address: draft.address, phone: draft.phone };
    try {
      if (draft.id) await api(`/portal/branches/${draft.id}`, { method: "PUT", body });
      else await api("/portal/branches", { method: "POST", body });
      setDraft(null);
      await load();
    } catch (err) {
      if (err instanceof ApiError) {
        setFields(err.fields);
        setError(Object.keys(err.fields).length ? "" : err.message);
      } else setError("Could not save the branch.");
    }
    setBusy(false);
  }

  async function remove(b: Branch) {
    if (!window.confirm(`Delete branch "${b.name}"?`)) return;
    setError("");
    try {
      await api(`/portal/branches/${b.id}`, { method: "DELETE" });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not delete the branch.");
    }
  }

  return (
    <>
      <div className="row">
        <h1>Branches</h1>
        <span className="spacer" />
        <button className="primary" onClick={() => { setFields({}); setError(""); setDraft({ ...empty }); }}>Add branch</button>
      </div>
      {error && !draft && <p className="error" role="alert">{error}</p>}

      {draft && (
        <form className="panel" onSubmit={save}>
          <h2>{draft.id ? "Edit branch" : "New branch"}</h2>
          <label>Name
            <input value={draft.name} onChange={(e) => setDraft({ ...draft, name: e.target.value })} required />
            {fields.name && <span className="error" role="alert">{fields.name}</span>}
          </label>
          <label>Area
            <select value={draft.area_id} onChange={(e) => setDraft({ ...draft, area_id: e.target.value })} required>
              <option value="">Choose an area</option>
              {areas.map((a) => <option key={a.id} value={a.id}>{a.name}</option>)}
            </select>
            {fields.area_id && <span className="error" role="alert">{fields.area_id}</span>}
          </label>
          <label>Address
            <input value={draft.address} onChange={(e) => setDraft({ ...draft, address: e.target.value })} required />
            {fields.address && <span className="error" role="alert">{fields.address}</span>}
          </label>
          <label>Phone (optional)
            <input value={draft.phone} onChange={(e) => setDraft({ ...draft, phone: e.target.value })} />
            {fields.phone && <span className="error" role="alert">{fields.phone}</span>}
          </label>
          {error && <p className="error" role="alert">{error}</p>}
          <div className="row">
            <button className="primary" disabled={busy}>{busy ? "Saving…" : "Save"}</button>
            <button type="button" onClick={() => setDraft(null)}>Cancel</button>
          </div>
        </form>
      )}

      {branches === null ? <p className="muted">Loading…</p> : branches.length === 0 ? (
        <p className="muted">No branches yet. Add your first branch before creating offers.</p>
      ) : (
        <table>
          <thead><tr><th>Name</th><th>Area</th><th>Address</th><th>Phone</th><th /></tr></thead>
          <tbody>
            {branches.map((b) => (
              <tr key={b.id}>
                <td>{b.name}</td><td>{b.area_name}</td><td>{b.address}</td><td>{b.phone ?? "—"}</td>
                <td className="actions">
                  <button onClick={() => edit(b)}>Edit</button>
                  <button className="danger" onClick={() => remove(b)}>Delete</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </>
  );
}
