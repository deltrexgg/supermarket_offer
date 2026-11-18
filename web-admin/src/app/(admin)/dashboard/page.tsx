"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { api, type Stats } from "@/lib/api";

export default function DashboardPage() {
  const [s, setS] = useState<Stats | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    api<Stats>("/admin/stats").then(setS).catch((e) => setError(e.message));
  }, []);

  if (error) return <p className="error">{error}</p>;
  if (!s) return <p className="muted">Loading…</p>;

  const tile = (label: string, value: number, href?: string, hint?: string) => {
    const body = (
      <>
        <span className="tile-value">{value}</span>
        <span className="tile-label">{label}</span>
        {hint && <span className="muted tile-hint">{hint}</span>}
      </>
    );
    return href ? <Link href={href} className="tile">{body}</Link> : <div className="tile">{body}</div>;
  };

  return (
    <>
      <h1>Dashboard</h1>
      {s.supermarkets_pending > 0 && (
        <p className="notice">
          {s.supermarkets_pending} supermarket{s.supermarkets_pending === 1 ? " is" : "s are"} waiting for approval.{" "}
          <Link href="/supermarkets?status=pending">Review now</Link>
        </p>
      )}
      <h2>Supermarkets</h2>
      <div className="grid">
        {tile("Pending approval", s.supermarkets_pending, "/supermarkets?status=pending")}
        {tile("Approved", s.supermarkets_approved, "/supermarkets?status=approved")}
        {tile("Blocked", s.supermarkets_blocked, "/supermarkets?status=blocked")}
        {tile("Branches", s.branches)}
      </div>
      <h2>Offers</h2>
      <div className="grid">
        {tile("Active today", s.offers_active, "/offers?status=active")}
        {tile("Ending within 2 days", s.offers_ending_soon, "/offers?status=active", "Counted from active offers")}
        {tile("Upcoming", s.offers_upcoming, "/offers?status=upcoming")}
        {tile("Expired", s.offers_expired, "/offers?status=expired")}
      </div>
    </>
  );
}
