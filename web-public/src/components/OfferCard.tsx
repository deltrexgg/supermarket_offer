import Link from "next/link";
import type { Card } from "@/lib/api";
import { branchesCount, daysLeftLabel, offerPath, pick, priceLabel, urgency, type Lang } from "@/lib/i18n";

export default function OfferCard({ o, lang }: { o: Card; lang: Lang }) {
  const name = pick(lang, o.item_name, o.item_name_ar);
  return (
    <Link href={offerPath(lang, o.id)} className="card">
      <div className="tile">
        {o.photo_url ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={o.photo_url} alt={name} loading="lazy" />
        ) : (
          <span className="tile-letter" aria-hidden="true">{name.charAt(0).toUpperCase()}</span>
        )}
        {o.discount_pct > 0 && <span className="off" dir="ltr">-{o.discount_pct}%</span>}
      </div>
      <div className="body">
        <span className="shop" dir="auto">{o.supermarket_name}</span>
        <h3 dir="auto">{name}</h3>
        <span className="muted" dir="auto">{[o.brand, o.unit].filter(Boolean).join(" · ")}</span>
        <div className="prices">
          <span className="now">{priceLabel(lang, o.offer_price)}</span>
          <span className="was">{priceLabel(lang, o.regular_price)}</span>
        </div>
        <div className="row">
          <span className={`chip ${urgency(o.days_left)}`}>{daysLeftLabel(lang, o.days_left)}</span>
          <span className="muted">{branchesCount(lang, o.branch_count)}</span>
        </div>
      </div>
    </Link>
  );
}
