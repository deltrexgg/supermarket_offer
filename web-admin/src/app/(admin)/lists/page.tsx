"use client";

import { useCallback, useEffect, useState } from "react";
import { api, type Lookup } from "@/lib/api";

export default function ListsPage() {
  return (
    <>
      <h1>Areas &amp; categories</h1>
      <p className="muted">These feed the filters on the public site and the forms in the supermarket portal.</p>
      <div className="two">
        <LookupPanel title="Areas" path="/admin/areas" usedBy="branches" />
        <LookupPanel title="Categories" path="/admin/categories" usedBy="offers" />
      </div>
    </>
  );
}

function LookupPanel({ title, path, usedBy }: { title: string; path: string; usedBy: string }) {
  const [items, setItems] = useState<Lookup[] | null>(null);
  const [name, setName] = useState("");
  const [editId, setEditId] = useState<number | null>(null);
  const [editName, setEditName] = useState("");
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    try {
      setItems((await api<{ items: Lookup[] }>(path)).items);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not load");
    }
  }, [path]);

  useEffect(() => {
    load();
  }, [load]);

  async function run(fn: () => Promise<unknown>) {
    setError("");
    try {
      await fn();
      await load();
    } catch (e) {
      const err = e as { message?: string; fields?: Record<string, string> };
      setError(err.fields?.name ?? err.message ?? "Failed");
    }
  }

  const add = (e: React.FormEvent) => {
    e.preventDefault();
    run(async () => {
      await api(path, { method: "POST", body: { name } });
      setName("");
    });
  };
  const save = (id: number) =>
    run(async () => {
      await api(`${path}/${id}`, { method: "PUT", body: { name: editName } });
      setEditId(null);
    });
  const remove = (l: Lookup) => {
    if (window.confirm(`Delete "${l.name}"?`)) run(() => api(`${path}/${l.id}`, { method: "DELETE" }));
  };

  return (
    <section className="panel">
      <h2>{title}</h2>
      <form className="row" onSubmit={add}>
        <input value={name} onChange={(e) => setName(e.target.value)} placeholder={`New ${title.toLowerCase().slice(0, -1)}`} aria-label={`New ${title}`} required minLength={2} maxLength={80} />
        <button className="primary">Add</button>
      </form>
      {error && <p className="error" role="alert">{error}</p>}
      {!items ? <p className="muted">Loading…</p> : (
        <table>
          <tbody>
            {items.map((l) => (
              <tr key={l.id}>
                <td>
                  {editId === l.id ? (
                    <input value={editName} onChange={(e) => setEditName(e.target.value)} aria-label="Name" autoFocus />
                  ) : l.name}
                </td>
                <td className="muted">{l.in_use} {usedBy}</td>
                <td className="actions">
                  {editId === l.id ? (
                    <>
                      <button className="primary" onClick={() => save(l.id)}>Save</button>
                      <button onClick={() => setEditId(null)}>Cancel</button>
                    </>
                  ) : (
                    <>
                      <button onClick={() => { setEditId(l.id); setEditName(l.name); }}>Rename</button>
                      <button className="danger" disabled={l.in_use > 0} title={l.in_use > 0 ? "In use" : undefined} onClick={() => remove(l)}>Delete</button>
                    </>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}
