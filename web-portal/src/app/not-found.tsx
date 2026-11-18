import Link from "next/link";

export default function NotFound() {
  return (
    <main className="auth-card">
      <h1>Page not found</h1>
      <p className="muted">That address does not exist.</p>
      <Link href="/">Go to the start</Link>
    </main>
  );
}
