import Link from "next/link";
import Icon, { categoryIcon } from "@/components/Icon";
import type { Option } from "@/lib/api";
import { home, strings, type Lang } from "@/lib/i18n";

// Category tiles; "All offers" first. The current one is filled green.
export default function Categories({ lang, categories, current }: { lang: Lang; categories: Option[]; current?: string }) {
  const t = strings(lang);
  if (categories.length === 0) return null;
  return (
    <section className="sec" aria-labelledby="cats-h">
      <h2 id="cats-h">{t.catTitle}</h2>
      <div className="cats">
        <Link href={`${home(lang)}#offers`} className="cat" aria-current={!current}>
          <Icon name="grid" size={28} />
          <span>{t.allOffers}</span>
        </Link>
        {categories.map((c) => (
          <Link key={c.id} href={`${home(lang)}?category=${c.id}#offers`} className="cat" aria-current={current === String(c.id)}>
            <Icon name={categoryIcon(c.name)} size={28} />
            <span dir="auto">{c.name}</span>
          </Link>
        ))}
      </div>
    </section>
  );
}
