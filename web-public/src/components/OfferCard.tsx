import Link from "next/link";
import type { Card } from "@/lib/api";
import { aed, daysLeftLabel, urgency } from "@/lib/format";

export default function OfferCard({ o }: { o: Card }) {
  return (
    <Link href={`/offers/${o.id}`} className="card">
      <div className="tile">
        {o.photo_url ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={o.photo_url} alt={o.item_name} loading="lazy" />
        ) : (
          <span className="tile-letter" aria-hidden="true">{o.item_name.charAt(0).toUpperCase()}</span>
        )}
        {o.discount_pct > 0 && <span className="off" dir="ltr">-{o.discount_pct}%</span>}
      </div>
      <div className="body">
        <span className="shop">{o.supermarket_name}</span>
        <h3>{o.item_name}</h3>
        <span className="muted">{[o.brand, o.unit].filter(Boolean).join(" · ")}</span>
        <div className="prices">
          <span className="now">{aed(o.offer_price)}</span>
          <span className="was">{aed(o.regular_price)}</span>
        </div>
        <div className="row">
          <span className={`chip ${urgency(o.days_left)}`}>{daysLeftLabel(o.days_left)}</span>
          <span className="muted">{o.branch_count} {o.branch_count === 1 ? "branch" : "branches"}</span>
        </div>
      </div>
    </Link>
  );
}
