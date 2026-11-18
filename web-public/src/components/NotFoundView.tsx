import Link from "next/link";
import { home, strings, type Lang } from "@/lib/i18n";

export default function NotFoundView({ lang, offer = false }: { lang: Lang; offer?: boolean }) {
  const t = strings(lang);
  return (
    <div className="empty">
      <h1>{offer ? t.notFoundTitle : t.pageNotFound}</h1>
      <p>
        {offer ? t.notFoundBody : t.pageNotFoundBody} <Link href={home(lang)}>{t.notFoundLink}</Link>
      </p>
    </div>
  );
}
