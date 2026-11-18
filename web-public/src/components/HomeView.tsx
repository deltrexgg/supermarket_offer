import Link from "next/link";
import { Suspense } from "react";
import Categories from "@/components/Categories";
import Filters from "@/components/Filters";
import Icon from "@/components/Icon";
import OfferCard from "@/components/OfferCard";
import TopOffers from "@/components/TopOffers";
import { getAreas, getCategories, getOffers, getSupermarkets } from "@/lib/api";
import { home, offersCount, pageLabel, strings, type Lang } from "@/lib/i18n";

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

const first = (v: string | string[] | undefined) => (Array.isArray(v) ? v[0] : v);

export default function HomeView({ lang, searchParams }: { lang: Lang; searchParams: SearchParams }) {
  const t = strings(lang);
  return (
    <div className="page">
      <h1 className="sr">{t.heroTitle}</h1>
      <Suspense fallback={<Loading label={t.loadingOffers} />}>
        <Results lang={lang} searchParams={searchParams} />
      </Suspense>
    </div>
  );
}

// Placeholder shapes shown while the offers load.
function Loading({ label }: { label: string }) {
  return (
    <>
      <p className="sr" role="status">{label}</p>
      <div className="grid" aria-hidden="true">
        {Array.from({ length: 8 }, (_, k) => (
          <div className="card sk" key={k}><div className="tile" /><div className="body"><i /><i /><i /></div></div>
        ))}
      </div>
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

  const filtered = Object.values(current).some(Boolean);
  // The slider shows the five biggest discounts on the unfiltered first page only.
  const showSlider = !filtered && page === 1;

  let data, areas, categories, supermarkets, top;
  try {
    [data, areas, categories, supermarkets, top] = await Promise.all([
      getOffers(qs), getAreas(), getCategories(), getSupermarkets(),
      showSlider
        ? getOffers(new URLSearchParams({ page: "1", page_size: "5", sort: "highest_discount" })).catch(() => null)
        : Promise.resolve(null),
    ]);
  } catch {
    return <p className="error">{t.unavailable}</p>;
  }

  const pages = Math.max(1, Math.ceil(data.total / data.page_size));
  const link = (p: number) => {
    const n = new URLSearchParams(qs);
    n.set("page", String(p));
    n.delete("page_size");
    return `${home(lang)}?${n}#offers`;
  };

  return (
    <>
      <div className="searchbar" role="search">
        <Icon name="search" />
        <input type="search" name="q" form="offerfilters" defaultValue={current.q ?? ""} placeholder={t.searchPlaceholder} aria-label={t.searchLabel} />
        <button type="submit" form="offerfilters" aria-label={t.search}><Icon name="search" /></button>
      </div>
      {top && top.items.length > 0 && <TopOffers offers={top.items} lang={lang} />}
      <Categories lang={lang} categories={categories} current={current.category} />
      <section className="sec" id="offers" aria-labelledby="offers-h">
        <div className="sech">
          <h2 id="offers-h">{t.offersTitle}</h2>
          <span>
            <span className="count">{offersCount(lang, data.total)}</span>
            {filtered && <> · <Link className="reset" href={home(lang)}>{t.clearFilters}</Link></>}
          </span>
        </div>
        <Filters lang={lang} areas={areas} categories={categories} supermarkets={supermarkets} current={current} />
        {data.items.length === 0 ? (
          <p className="empty">
            <span>{t.emptyBefore}<Link href={home(lang)}>{t.emptyLink}</Link>{t.emptyAfter}</span>
          </p>
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
      </section>
    </>
  );
}
