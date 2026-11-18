import Link from "next/link";
import Icon, { CAT_TINTS, categoryIcon } from "@/components/Icon";
import type { Option } from "@/lib/api";
import { home, strings, type Lang } from "@/lib/i18n";

// Tappable category tiles; each one applies the category filter.
export default function Categories({ lang, categories, current }: { lang: Lang; categories: Option[]; current?: string }) {
  const t = strings(lang);
  if (categories.length === 0) return null;
  return (
    <section className="sec" aria-labelledby="cats-h">
      <h2 id="cats-h">{t.catTitle}</h2>
      <div className="cats">
        {categories.map((c, idx) => (
          <Link key={c.id} href={`${home(lang)}?category=${c.id}#offers`} className="cat" aria-current={current === String(c.id)}>
            <span className="caticon" style={{ background: CAT_TINTS[idx % CAT_TINTS.length] }}><Icon name={categoryIcon(c.name)} size={24} /></span>
            <span className="catn" dir="auto">{c.name}</span>
          </Link>
        ))}
      </div>
    </section>
  );
}
