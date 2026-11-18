import Link from "next/link";
import { Suspense } from "react";
import Icon, { categoryIcon } from "@/components/Icon";
import Logo from "@/components/Logo";
import OfferCard from "@/components/OfferCard";
import TopOffers from "@/components/TopOffers";
import { getAreas, getCategories, getOffers, getSupermarkets, type Option, type Shop } from "@/lib/api";
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

// Placeholder shapes while the offers load.
function Loading({ label }: { label: string }) {
  return (
    <>
      <p className="sr" role="status">{label}</p>
      <div className="grid" aria-hidden="true">
        {Array.from({ length: 8 }, (_, k) => (
          <div className="card sk" key={k}><div className="tile" /><i /><i /><i /></div>
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

  let data, areas: Option[], categories: Option[], supermarkets: Shop[], top;
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

  // Every chip, tab and store keeps the other filters and starts again at page 1.
  const linkWith = (key: string, value?: string) => {
    const n = new URLSearchParams();
    for (const [k, v] of Object.entries(current)) if (v && k !== key) n.set(k, v);
    if (value) n.set(key, value);
    const s = n.toString();
    return `${home(lang)}${s ? `?${s}` : ""}#offers`;
  };
  const pageLink = (p: number) => {
    const n = new URLSearchParams(qs);
    n.set("page", String(p));
    n.delete("page_size");
    return `${home(lang)}?${n}#offers`;
  };
  const pages = Math.max(1, Math.ceil(data.total / data.page_size));

  const sorts: [string, string][] = [
    ["highest_discount", t.topPicks],
    ["ending_soon", t.endingSoon],
    ["lowest_price", t.lowestPrice],
  ];
  const sort = current.sort ?? "highest_discount";

  return (
    <>
      <form method="get" action={home(lang)} className="sbar">
        <label className="sfield">
          <span className="sr">{t.searchLabel}</span>
          <Icon name="search" size={18} />
          <input type="search" name="q" defaultValue={current.q ?? ""} placeholder={t.searchPlaceholder} />
        </label>
        <span className="areapick">
          <label htmlFor="area">{t.area}</label>
          <select id="area" name="area" defaultValue={current.area ?? ""}>
            <option value="">{t.all}</option>
            {areas.map((a) => <option key={a.id} value={a.id}>{a.name}</option>)}
          </select>
        </span>
        <button className="sgo" type="submit">{t.search}</button>
      </form>
      <div className="rule" />

      <nav className="strip" aria-label={t.storesLabel}>
        <Link className="store" href={linkWith("supermarket")} aria-current={!current.supermarket}>{t.allStores}</Link>
        {supermarkets.map((s) => (
          <Link key={s.id} className="store" href={linkWith("supermarket", String(s.id))} aria-current={current.supermarket === String(s.id)}>
            <Logo name={s.name} url={s.logo_url} />
            <span dir="auto">{s.name}</span>
          </Link>
        ))}
      </nav>
      <div className="rule" style={{ marginBottom: 22 }} />

      {top && top.items.length > 0 && <TopOffers offers={top.items} lang={lang} />}

      <nav className="chips" aria-label={t.catTitle}>
        <Link className="chip" href={linkWith("category")} aria-current={!current.category}>
          <Icon name="grid" size={17} /> {t.allOffers}
        </Link>
        {categories.map((c) => (
          <Link key={c.id} className="chip" href={linkWith("category", String(c.id))} aria-current={current.category === String(c.id)}>
            <Icon name={categoryIcon(c.name)} size={17} />
            <span dir="auto">{c.name}</span>
          </Link>
        ))}
      </nav>

      <section id="offers" aria-labelledby="offers-h">
        <div className="sech">
          <h2 className="h2" id="offers-h">
            {t.offersTitle} <span className="cnt">{offersCount(lang, data.total)}</span>
            {filtered && <Link className="reset" href={home(lang)}>{t.clearFilters}</Link>}
          </h2>
          <nav className="tabs" aria-label={t.sortBy}>
            {sorts.map(([value, label]) => (
              <Link key={value} className="tab" href={linkWith("sort", value)} aria-current={sort === value}>{label}</Link>
            ))}
          </nav>
        </div>
        {data.items.length === 0 ? (
          <p className="empty">
            <span>{t.emptyBefore}<Link href={home(lang)}>{t.emptyLink}</Link>{t.emptyAfter}</span>
          </p>
        ) : (
          <div className="grid">
            {data.items.map((o, k) => <OfferCard key={o.id} o={o} lang={lang} index={k} />)}
          </div>
        )}
        {pages > 1 && (
          <nav className="pager" aria-label={t.pagesLabel}>
            {page > 1 ? <Link href={pageLink(page - 1)}>{t.previous}</Link> : <span />}
            <span className="muted">{pageLabel(lang, page, pages)}</span>
            {page < pages ? <Link href={pageLink(page + 1)}>{t.next}</Link> : <span />}
          </nav>
        )}
      </section>
    </>
  );
}
