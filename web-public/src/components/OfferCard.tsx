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
        {o.discount_pct > 0 && <span className="off"><bdi dir="ltr">-{o.discount_pct}%</bdi></span>}
      </div>
      <div className="body">
        <h3 dir="auto">{name}</h3>
        <span className="sub" dir="auto">
          {o.logo_url ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img className="shoplogo" src={o.logo_url} alt="" loading="lazy" />
          ) : (
            <span className="shoplogo letter" aria-hidden="true">{o.supermarket_name.charAt(0).toUpperCase()}</span>
          )}
          {[o.supermarket_name, o.brand, o.unit].filter(Boolean).join(" · ")}
        </span>
        <span className={`chip ${urgency(o.days_left)}`}>{daysLeftLabel(lang, o.days_left)}</span>
        <div className="prow">
          <div className="prices">
            <span className="now">{priceLabel(lang, o.offer_price)}</span>
            <span className="was">{priceLabel(lang, o.regular_price)}</span>
          </div>
          <span className="go" aria-hidden="true"><Icon name="arrow" size={18} /></span>
        </div>
      </div>
    </Link>
  );
}
