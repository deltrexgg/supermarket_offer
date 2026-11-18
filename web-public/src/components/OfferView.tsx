import Link from "next/link";
import { notFound } from "next/navigation";
import { Suspense } from "react";
import Icon from "@/components/Icon";
import { getOffer, NotFoundError } from "@/lib/api";
import {
  branchesCount, daysLeftLabel, home, pick, priceLabel, saveLabel, strings, urgency, validRange, type Lang,
} from "@/lib/i18n";

export default function OfferView({ lang, params }: { lang: Lang; params: Promise<{ id: string }> }) {
  const t = strings(lang);
  return (
    <div className="page">
      <Link href={home(lang)} className="back"><span className="flip" style={{ display: "inline-flex" }}><Icon name="back" size={18} /></span> {t.back}</Link>
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
        <div className="photo">
          {o.photo_url ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={o.photo_url} alt={name} />
          ) : (
            <span className="tile-letter" aria-hidden="true">{name.charAt(0).toUpperCase()}</span>
          )}
          {o.discount_pct > 0 && <span className="off"><bdi dir="ltr">-{o.discount_pct}%</bdi></span>}
        </div>
        <div className="panel">
          <div className="tags">
            <span className="tag" dir="auto">
              {o.logo_url ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img className="shoplogo" src={o.logo_url} alt="" />
              ) : (
                <span className="shoplogo letter" aria-hidden="true">{o.supermarket_name.charAt(0).toUpperCase()}</span>
              )}
              {o.supermarket_name}
            </span>
            <span className="tag n" dir="auto">{o.category_name}</span>
          </div>
          <h1 dir="auto">{name}</h1>
          {(o.brand || o.unit) && <p className="muted" dir="auto" style={{ margin: 0 }}>{[o.brand, o.unit].filter(Boolean).join(" · ")}</p>}
          <div className="big">
            <span className="now">{priceLabel(lang, o.offer_price)}</span>
            <span className="was">{priceLabel(lang, o.regular_price)}</span>
          </div>
          {saves && <span className="saveb">{saveLabel(lang, o.regular_price, o.offer_price)}</span>}
          <div className="valid">
            <Icon name="calendar" />
            <span className="vtxt">{validRange(lang, o.valid_from, o.valid_to)}</span>
            <span className={`chip ${urgency(o.days_left)}`}>{daysLeftLabel(lang, o.days_left)}</span>
          </div>
          {description && <p className="desc" dir="auto">{description}</p>}
        </div>
      </article>
      <section className="sec" aria-labelledby="br-h">
        <div className="sech">
          <h2 id="br-h">{t.branchesTitle}</h2>
          <span className="count">{branchesCount(lang, o.branches.length)}</span>
        </div>
        <ul className="branches">
          {o.branches.map((b) => {
            const phone = b.phone ?? o.supermarket_phone;
            return (
              <li key={b.id}>
                <strong dir="auto">{b.name}</strong>
                <span className="brarea" dir="auto">{b.area_name}</span>
                <span className="muted" dir="auto" style={{ fontSize: 14 }}>{b.address}</span>
                {phone && (
                  <a className="call" href={`tel:${phone.replace(/\s+/g, "")}`}>
                    <Icon name="phone" size={18} />
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
