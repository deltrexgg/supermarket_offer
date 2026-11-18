import Link from "next/link";
import Logo from "@/components/Logo";
import type { Card } from "@/lib/api";
import { daysLeftLabel, offerPath, offLabel, pick, priceLabel, saveLabel, strings, urgency, type Lang } from "@/lib/i18n";

// Tints cycle down the grid so each offer reads as its own thing.
const TINTS = ["var(--tint-1)", "var(--tint-2)", "var(--tint-3)", "var(--tint-4)", "var(--tint-5)"];

export default function OfferCard({ o, lang, index }: { o: Card; lang: Lang; index: number }) {
  const t = strings(lang);
  const name = pick(lang, o.item_name, o.item_name_ar);
  const saves = o.regular_price > o.offer_price;
  return (
    <Link href={offerPath(lang, o.id)} className="card">
      <div className="tile" style={{ background: TINTS[index % TINTS.length] }}>
        {o.photo_url ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={o.photo_url} alt={name} loading="lazy" />
        ) : (
          <span className="tile-letter" aria-hidden="true">{name.charAt(0).toUpperCase()}</span>
        )}
        {o.discount_pct > 0 && (
          <span className="disc">
            <span className="pctn">{offLabel(lang, o.discount_pct)}</span>
            <span className="pctw">{t.offWord}</span>
          </span>
        )}
      </div>
      <span className="shoptag" dir="auto">
        <Logo name={o.supermarket_name} url={o.logo_url} />
        {o.supermarket_name}
      </span>
      <h3 className="nm" dir="auto">{name}</h3>
      <p className="unit" dir="auto">{[o.brand, o.unit].filter(Boolean).join(" · ")}</p>
      <div className="prule" />
      <div className="prow">
        <span className="now">{priceLabel(lang, o.offer_price)}</span>
        <span className="was">{priceLabel(lang, o.regular_price)}</span>
      </div>
      {saves && <span className="savetag">{saveLabel(lang, o.regular_price, o.offer_price)}</span>}
      <span className={`dleft ${urgency(o.days_left)}`}>
        <svg className="dot" viewBox="0 0 8 8" aria-hidden="true"><circle cx="4" cy="4" r="4" fill="currentColor" /></svg>
        {daysLeftLabel(lang, o.days_left)}
      </span>
    </Link>
  );
}
