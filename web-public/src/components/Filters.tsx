import type { Option } from "@/lib/api";
import { home, strings, type Lang } from "@/lib/i18n";

// A plain GET form: it works without JavaScript and keeps every filter in the URL.
export default function Filters({
  lang, areas, categories, supermarkets, current,
}: {
  lang: Lang; areas: Option[]; categories: Option[]; supermarkets: Option[];
  current: { q?: string; area?: string; category?: string; supermarket?: string; sort?: string };
}) {
  const t = strings(lang);
  const select = (name: string, label: string, options: Option[], value?: string) => (
    <div className="field">
      <label htmlFor={name}>{label}</label>
      <select id={name} name={name} defaultValue={value ?? ""}>
        <option value="">{t.all}</option>
        {options.map((o) => <option key={o.id} value={o.id}>{o.name}</option>)}
      </select>
    </div>
  );
  return (
    <form method="get" action={home(lang)} className="filterform">
      <div className="search">
        <input type="search" name="q" defaultValue={current.q ?? ""} placeholder={t.searchPlaceholder} aria-label={t.searchLabel} />
        <button className="primary" type="submit">{t.search}</button>
      </div>
      <div className="filters">
        {select("area", t.area, areas, current.area)}
        {select("category", t.category, categories, current.category)}
        {select("supermarket", t.supermarket, supermarkets, current.supermarket)}
        <div className="field">
          <label htmlFor="sort">{t.sortBy}</label>
          <select id="sort" name="sort" defaultValue={current.sort ?? "highest_discount"}>
            <option value="highest_discount">{t.highestDiscount}</option>
            <option value="lowest_price">{t.lowestPrice}</option>
            <option value="ending_soon">{t.endingSoon}</option>
          </select>
        </div>
      </div>
    </form>
  );
}
