import Link from "next/link";
import { notFound } from "next/navigation";
import { Suspense } from "react";
import { getOffer, NotFoundError } from "@/lib/api";
import { aed, daysLeftLabel, longDate, urgency } from "@/lib/format";

export default function OfferPage({ params }: { params: Promise<{ id: string }> }) {
  return (
    <>
      <p><Link href="/" className="back">← All offers</Link></p>
      <Suspense fallback={<p className="muted">Loading offer…</p>}>
        <Detail params={params} />
      </Suspense>
    </>
  );
}

async function Detail({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  let o;
  try {
    o = await getOffer(id);
  } catch (e) {
    if (e instanceof NotFoundError) notFound();
    return <p className="error">This offer can&apos;t be loaded right now. Please try again in a moment.</p>;
  }

  return (
    <article className="detail">
      <div className="photo">
        {o.photo_url ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={o.photo_url} alt={o.item_name} />
        ) : (
          <span className="tile-letter" aria-hidden="true">{o.item_name.charAt(0).toUpperCase()}</span>
        )}
        {o.discount_pct > 0 && <span className="off" dir="ltr">-{o.discount_pct}%</span>}
      </div>
      <div className="info">
        <span className="shop">{o.supermarket_name}</span>
        <h1>{o.item_name}</h1>
        <p className="muted">{[o.brand, o.unit, o.category_name].filter(Boolean).join(" · ")}</p>
        <div className="prices big">
          <span className="now">{aed(o.offer_price)}</span>
          <span className="was">{aed(o.regular_price)}</span>
        </div>
        <p>
          <span className={`chip ${urgency(o.days_left)}`}>{daysLeftLabel(o.days_left)}</span>{" "}
          <span className="muted">Valid {longDate(o.valid_from)} to {longDate(o.valid_to)}</span>
        </p>
        {o.description && <p>{o.description}</p>}

        <h2>Available at</h2>
        <ul className="branches">
          {o.branches.map((b) => (
            <li key={b.id}>
              <strong>{b.name}</strong> <span className="muted">· {b.area_name}</span>
              <br />
              <span>{b.address}</span>
              {(b.phone || o.supermarket_phone) && (
                <>
                  <br />
                  <a href={`tel:${(b.phone ?? o.supermarket_phone ?? "").replace(/\s+/g, "")}`}>{b.phone ?? o.supermarket_phone}</a>
                </>
              )}
            </li>
          ))}
        </ul>
      </div>
    </article>
  );
}
