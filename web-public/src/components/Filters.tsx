"use client";

import { useState } from "react";
import Icon from "@/components/Icon";
import type { Option } from "@/lib/api";
import { useHydrated } from "@/lib/hooks";
import { home, strings, type Lang } from "@/lib/i18n";

// A plain GET form: it works without JavaScript and keeps every filter in the URL.
// With JavaScript, phones get a Filters button that folds the fields away.
// The search box lives at the top of the page and joins this form through its id.
export default function Filters({
  lang, areas, categories, supermarkets, current,
}: {
  lang: Lang; areas: Option[]; categories: Option[]; supermarkets: Option[];
  current: { area?: string; category?: string; supermarket?: string; sort?: string };
}) {
  const t = strings(lang);
  const js = useHydrated();
  const [open, setOpen] = useState(false);

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
    <>
      {js && (
        <button type="button" className="filterbtn" aria-expanded={open} aria-controls="filter-fields" onClick={() => setOpen((v) => !v)}>
          <Icon name="filter" size={18} /> {t.filters}
        </button>
      )}
      <form id="offerfilters" method="get" action={home(lang)}>
        <div id="filter-fields" className={`filters${js && !open ? " collapsed" : ""}`}>
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
          <button className="primary" type="submit">{t.apply}</button>
        </div>
      </form>
    </>
  );
}
