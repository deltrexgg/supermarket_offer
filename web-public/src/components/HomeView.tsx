import Link from "next/link";
import { Suspense } from "react";
import Filters from "@/components/Filters";
import OfferCard from "@/components/OfferCard";
import { getAreas, getCategories, getOffers, getSupermarkets } from "@/lib/api";
import { home, offersCount, pageLabel, strings, type Lang } from "@/lib/i18n";

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

const first = (v: string | string[] | undefined) => (Array.isArray(v) ? v[0] : v);

export default function HomeView({ lang, searchParams }: { lang: Lang; searchParams: SearchParams }) {
  const t = strings(lang);
  return (
    <>
      <section className="hero">
        <h1>{t.heroTitle}</h1>
        <p className="muted">{t.heroSub}</p>
      </section>
      <Suspense fallback={<p className="muted">{t.loadingOffers}</p>}>
        <Results lang={lang} searchParams={searchParams} />
      </Suspense>
    </>
  );
}

async function Results({ lang, searchParams }: { lang: Lang; searchParams: SearchParams }) {
  const t = strings(lang);
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
    return <p className="error">{t.unavailable}</p>;
  }

  const pages = Math.max(1, Math.ceil(data.total / data.page_size));
  const link = (p: number) => {
    const n = new URLSearchParams(qs);
    n.set("page", String(p));
    n.delete("page_size");
    return `${home(lang)}?${n}`;
  };
  const filtered = Object.values(current).some(Boolean);

  return (
    <>
      <Filters lang={lang} areas={areas} categories={categories} supermarkets={supermarkets} current={current} />
      <div className="bar">
        <span className="count">{offersCount(lang, data.total)}</span>
        {filtered && <Link className="reset" href={home(lang)}>{t.clearFilters}</Link>}
      </div>
      {data.items.length === 0 ? (
        <p className="empty">{t.emptyBefore}<Link href={home(lang)}>{t.emptyLink}</Link>{t.emptyAfter}</p>
      ) : (
        <div className="grid">{data.items.map((o) => <OfferCard key={o.id} o={o} lang={lang} />)}</div>
      )}
      {pages > 1 && (
        <nav className="pager" aria-label={t.pagesLabel}>
          {page > 1 ? <Link href={link(page - 1)}>{t.previous}</Link> : <span />}
          <span className="muted">{pageLabel(lang, page, pages)}</span>
          {page < pages ? <Link href={link(page + 1)}>{t.next}</Link> : <span />}
        </nav>
      )}
    </>
  );
}
