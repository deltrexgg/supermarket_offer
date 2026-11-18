import Link from "next/link";

export default function NotFound() {
  return (
    <div className="empty">
      <h1>Page not found</h1>
      <p>That address does not exist. <Link href="/">See today&apos;s offers</Link>.</p>
    </div>
  );
}
