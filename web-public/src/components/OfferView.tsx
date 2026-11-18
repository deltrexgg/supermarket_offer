import Link from "next/link";
import { notFound } from "next/navigation";
import { Suspense } from "react";
import { getOffer, NotFoundError } from "@/lib/api";
import {
  daysLeftLabel, home, pick, priceLabel, strings, urgency, validRange, type Lang,
} from "@/lib/i18n";

export default function OfferView({ lang, params }: { lang: Lang; params: Promise<{ id: string }> }) {
  const t = strings(lang);
  return (
    <>
      <p><Link href={home(lang)} className="back">{t.back}</Link></p>
      <Suspense fallback={<p className="muted">{t.loadingOffer}</p>}>
        <Detail lang={lang} params={params} />
      </Suspense>
    </>
  );
}

async function Detail({ lang, params }: { lang: Lang; params: Promise<{ id: string }> }) {
  const t = strings(lang);
  const { id } = await params;
  let o;
  try {
    o = await getOffer(id);
  } catch (e) {
    if (e instanceof NotFoundError) notFound();
    return <p className="error">{t.offerUnavailable}</p>;
  }
  const name = pick(lang, o.item_name, o.item_name_ar);
  const description = pick(lang, o.description, o.description_ar);

  return (
    <article className="detail">
      <div className="photo">
        {o.photo_url ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={o.photo_url} alt={name} />
        ) : (
          <span className="tile-letter" aria-hidden="true">{name.charAt(0).toUpperCase()}</span>
        )}
        {o.discount_pct > 0 && <span className="off" dir="ltr">-{o.discount_pct}%</span>}
      </div>
      <div className="info">
        <span className="shop" dir="auto">{o.supermarket_name}</span>
        <h1 dir="auto">{name}</h1>
        <p className="muted" dir="auto">{[o.brand, o.unit, o.category_name].filter(Boolean).join(" · ")}</p>
        <div className="prices big">
          <span className="now">{priceLabel(lang, o.offer_price)}</span>
          <span className="was">{priceLabel(lang, o.regular_price)}</span>
        </div>
        <p>
          <span className={`chip ${urgency(o.days_left)}`}>{daysLeftLabel(lang, o.days_left)}</span>{" "}
          <span className="muted">{validRange(lang, o.valid_from, o.valid_to)}</span>
        </p>
        {description && <p dir="auto">{description}</p>}

        <h2>{t.availableAt}</h2>
        <ul className="branches">
          {o.branches.map((b) => (
            <li key={b.id}>
              <strong dir="auto">{b.name}</strong> <span className="muted" dir="auto">· {b.area_name}</span>
              <br />
              <span dir="auto">{b.address}</span>
              {(b.phone || o.supermarket_phone) && (
                <>
                  <br />
                  <a href={`tel:${(b.phone ?? o.supermarket_phone ?? "").replace(/\s+/g, "")}`} dir="ltr">{b.phone ?? o.supermarket_phone}</a>
                </>
              )}
            </li>
          ))}
        </ul>
      </div>
    </article>
  );
}
