import Link from "next/link";

export default function NotFound() {
  return (
    <div className="empty">
      <h1>This offer isn&apos;t available</h1>
      <p>It may have ended, or it isn&apos;t published yet. <Link href="/">See today&apos;s offers</Link>.</p>
    </div>
  );
}
