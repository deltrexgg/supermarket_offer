import Link from "next/link";
import { Suspense } from "react";
import Filters from "@/components/Filters";
import OfferCard from "@/components/OfferCard";
import { getAreas, getCategories, getOffers, getSupermarkets } from "@/lib/api";

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

const first = (v: string | string[] | undefined) => (Array.isArray(v) ? v[0] : v);

export default function Home({ searchParams }: { searchParams: SearchParams }) {
  return (
    <>
      <section className="hero">
        <h1>Today&apos;s supermarket offers in Dubai</h1>
        <p className="muted">Compare live deals from supermarkets near you. Prices in AED.</p>
      </section>
      <Suspense fallback={<p className="muted">Loading offers…</p>}>
        <Results searchParams={searchParams} />
      </Suspense>
    </>
  );
}

async function Results({ searchParams }: { searchParams: SearchParams }) {
  const sp = await searchParams;
  const current = {
    q: first(sp.q), area: first(sp.area), category: first(sp.category),
    supermarket: first(sp.supermarket), sort: first(sp.sort),
  };
  const page = Math.max(1, Number(first(sp.page)) || 1);

  const qs = new URLSearchParams({ page: String(page), page_size: "12" });
  for (const [k, v] of Object.entries(current)) if (v) qs.set(k, v);

  let data, areas, categories, supermarkets;
  try {
    [data, areas, categories, supermarkets] = await Promise.all([
      getOffers(qs), getAreas(), getCategories(), getSupermarkets(),
    ]);
  } catch {
    return <p className="error">Offers are unavailable right now. Please try again in a moment.</p>;
  }

  const pages = Math.max(1, Math.ceil(data.total / data.page_size));
  const link = (p: number) => {
    const n = new URLSearchParams(qs);
    n.set("page", String(p));
    n.delete("page_size");
    return `/?${n}`;
  };
  const filtered = Object.values(current).some(Boolean);

  return (
    <>
      <Filters areas={areas} categories={categories} supermarkets={supermarkets} current={current} />
      <div className="bar">
        <span className="count">{data.total} {data.total === 1 ? "offer" : "offers"}</span>
        {filtered && <Link className="reset" href="/">Clear filters</Link>}
      </div>
      {data.items.length === 0 ? (
        <p className="empty">No offers match your search. Try a different word, or <Link href="/">clear the filters</Link>.</p>
      ) : (
        <div className="grid">{data.items.map((o) => <OfferCard key={o.id} o={o} />)}</div>
      )}
      {pages > 1 && (
        <nav className="pager" aria-label="Pages">
          {page > 1 ? <Link href={link(page - 1)}>Previous</Link> : <span />}
          <span className="muted">Page {page} of {pages}</span>
          {page < pages ? <Link href={link(page + 1)}>Next</Link> : <span />}
        </nav>
      )}
    </>
  );
}
