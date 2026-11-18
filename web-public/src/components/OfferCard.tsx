import Link from "next/link";
import Icon from "@/components/Icon";
import type { Card } from "@/lib/api";
import { daysLeftLabel, offerPath, pick, priceLabel, urgency, type Lang } from "@/lib/i18n";

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
        <span className="shoptag" dir="auto">{o.supermarket_name}</span>
      </div>
      <div className="body">
        <h3 dir="auto">{name}</h3>
        <span className="muted" dir="auto" style={{ fontSize: 13 }}>{[o.brand, o.unit].filter(Boolean).join(" · ")}</span>
        <div className="prow">
          <div className="prices">
            <span className="now">{priceLabel(lang, o.offer_price)}</span>
            <span className="was">{priceLabel(lang, o.regular_price)}</span>
          </div>
          <span className="go" aria-hidden="true"><Icon name="arrow" /></span>
        </div>
        <span className={`chip ${urgency(o.days_left)}`}>{daysLeftLabel(lang, o.days_left)}</span>
      </div>
    </Link>
  );
}
