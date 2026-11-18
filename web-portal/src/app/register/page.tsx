"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { api, ApiError } from "@/lib/api";
import { supabase } from "@/lib/supabase";

export default function RegisterPage() {
  const router = useRouter();
  const [form, setForm] = useState({ name: "", email: "", password: "", phone: "" });
  const [fields, setFields] = useState<Record<string, string>>({});
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const set = (k: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm({ ...form, [k]: e.target.value });

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    setFields({});
    try {
      await api("/auth/register", { method: "POST", body: form, auth: false });
      const { error } = await supabase().auth.signInWithPassword({ email: form.email.trim(), password: form.password });
      if (error) throw new Error("Account created. Please sign in.");
      router.replace("/offers");
    } catch (err) {
      if (err instanceof ApiError) {
        setFields(err.fields);
        setError(Object.keys(err.fields).length ? "" : err.message);
      } else {
        setError(err instanceof Error ? err.message : "Something went wrong.");
      }
      setBusy(false);
    }
  }

  return (
    <main className="auth-card">
      <h1>Create your account</h1>
      <p className="muted">Register your supermarket to publish offers.</p>
      <form onSubmit={submit}>
        <label>Supermarket name
          <input value={form.name} onChange={set("name")} required />
          {fields.name && <span className="error">{fields.name}</span>}
        </label>
        <label>Email
          <input type="email" value={form.email} onChange={set("email")} required autoComplete="email" />
          {fields.email && <span className="error">{fields.email}</span>}
        </label>
        <label>Password (8+ characters)
          <input type="password" value={form.password} onChange={set("password")} required autoComplete="new-password" />
          {fields.password && <span className="error">{fields.password}</span>}
        </label>
        <label>Phone (optional)
          <input value={form.phone} onChange={set("phone")} autoComplete="tel" />
          {fields.phone && <span className="error">{fields.phone}</span>}
        </label>
        {error && <p className="error">{error}</p>}
        <button className="primary" disabled={busy}>{busy ? "Creating…" : "Create account"}</button>
      </form>
      <p className="muted">Already registered? <Link href="/login">Sign in</Link></p>
    </main>
  );
}
