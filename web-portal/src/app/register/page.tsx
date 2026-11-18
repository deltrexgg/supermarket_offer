"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { supabase } from "@/lib/supabase";

export default function RegisterPage() {
  const router = useRouter();
  const [form, setForm] = useState({ name: "", email: "", password: "", phone: "" });
  const [logo, setLogo] = useState<File | null>(null);
  const [fields, setFields] = useState<Record<string, string>>({});
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const logoPreview = useMemo(() => (logo ? URL.createObjectURL(logo) : null), [logo]);

  const set = (k: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm({ ...form, [k]: e.target.value });

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    setFields({});
    try {
      // With a logo the signup goes up as multipart; without one it stays JSON.
      let body: unknown = form;
      if (logo) {
        const fd = new FormData();
        for (const [k, v] of Object.entries(form)) fd.append(k, v);
        fd.append("logo", logo);
        body = fd;
      }
      await api("/auth/register", { method: "POST", body, auth: false });
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
          {fields.name && <span className="error" role="alert">{fields.name}</span>}
        </label>
        <label>Email
          <input type="email" value={form.email} onChange={set("email")} required autoComplete="email" />
          {fields.email && <span className="error" role="alert">{fields.email}</span>}
        </label>
        <label>Password (8+ characters)
          <input type="password" value={form.password} onChange={set("password")} required autoComplete="new-password" />
          {fields.password && <span className="error" role="alert">{fields.password}</span>}
        </label>
        <label>Phone (optional)
          <input value={form.phone} onChange={set("phone")} autoComplete="tel" />
          {fields.phone && <span className="error" role="alert">{fields.phone}</span>}
        </label>
        <label>Logo (optional)
          <input type="file" accept="image/jpeg,image/png,image/webp,image/gif,image/avif" onChange={(e) => setLogo(e.target.files?.[0] ?? null)} />
          <span className="muted">Shoppers see this next to your name. JPEG, PNG, WebP, GIF or AVIF, up to 5 MB. A square image works best.</span>
          {fields.logo && <span className="error" role="alert">{fields.logo}</span>}
        </label>
        {logoPreview && (
          // eslint-disable-next-line @next/next/no-img-element
          <img className="logo-preview" src={logoPreview} alt="Logo preview" />
        )}
        {error && <p className="error" role="alert">{error}</p>}
        <button className="primary" disabled={busy}>{busy ? "Creating…" : "Create account"}</button>
      </form>
      <p className="muted">Already registered? <Link href="/login">Sign in</Link></p>
    </main>
  );
}
