import Link from "next/link";
import { notFound } from "next/navigation";
import { Suspense } from "react";
import Icon from "@/components/Icon";
import Logo from "@/components/Logo";
import { getOffer, NotFoundError } from "@/lib/api";
import {
  branchesCount, daysLeftLabel, home, offLabel, pick, priceLabel, saveLabel, strings, urgency, validRange, type Lang,
} from "@/lib/i18n";

export default function OfferView({ lang, params }: { lang: Lang; params: Promise<{ id: string }> }) {
  const t = strings(lang);
  return (
    <div className="page">
      <Link href={home(lang)} className="back">
        <span className="flip" style={{ display: "inline-flex" }}><Icon name="back" size={16} /></span> {t.back}
      </Link>
      <Suspense fallback={<p className="muted" role="status">{t.loadingOffer}</p>}>
        <Detail lang={lang} params={params} />
      </Suspense>
    </div>
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
  const saves = o.regular_price > o.offer_price;

  return (
    <>
      <article className="detail">
        <div className="hero">
          {o.photo_url ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={o.photo_url} alt={name} />
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
        <div className="panel">
          <span className="eyebrow" dir="auto">
            <Logo name={o.supermarket_name} url={o.logo_url} />
            {[o.supermarket_name, o.category_name].filter(Boolean).join(" · ")}
          </span>
          <h1 className="h1" dir="auto">{name}</h1>
          {(o.brand || o.unit) && <p className="unit" dir="auto">{[o.brand, o.unit].filter(Boolean).join(" · ")}</p>}
          <div className="prule" />
          <div className="prow">
            <span className="big">{priceLabel(lang, o.offer_price)}</span>
            <span className="was">{priceLabel(lang, o.regular_price)}</span>
          </div>
          {saves && <span className="saveb">{saveLabel(lang, o.regular_price, o.offer_price)}</span>}
          <div className="valid">
            <Icon name="calendar" size={17} />
            <span className="vtxt">{validRange(lang, o.valid_from, o.valid_to)}</span>
            <span className={`dleft ${urgency(o.days_left)}`}>
              <svg className="dot" viewBox="0 0 8 8" aria-hidden="true"><circle cx="4" cy="4" r="4" fill="currentColor" /></svg>
              {daysLeftLabel(lang, o.days_left)}
            </span>
          </div>
          {description && <p className="desc" dir="auto">{description}</p>}
        </div>
      </article>
      <section aria-labelledby="br-h">
        <div className="sech">
          <h2 className="h2" id="br-h">{t.branchesTitle}</h2>
          <span className="cnt">{branchesCount(lang, o.branches.length)}</span>
        </div>
        <ul className="branches">
          {o.branches.map((b) => {
            const phone = b.phone ?? o.supermarket_phone;
            return (
              <li key={b.id}>
                <span className="brn" dir="auto">{b.name}</span>
                <span className="brarea" dir="auto">{b.area_name}</span>
                <span className="brm" dir="auto"><Icon name="pin" size={16} />{b.address}</span>
                {phone && (
                  <a className="call" href={`tel:${phone.replace(/\s+/g, "")}`}>
                    <Icon name="phone" size={16} />
                    <span dir="ltr">{phone}</span>
                  </a>
                )}
              </li>
            );
          })}
        </ul>
      </section>
    </>
  );
}
