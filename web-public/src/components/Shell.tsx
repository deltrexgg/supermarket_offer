import Link from "next/link";
import { Suspense } from "react";
import LangSwitch from "@/components/LangSwitch";
import { home, strings, type Lang } from "@/lib/i18n";

// Masthead, content and footer shared by the English and Arabic sites.
export default function Shell({ lang, children }: { lang: Lang; children: React.ReactNode }) {
  const t = strings(lang);
  return (
    <div className="wrap">
      <header className="mast">
        <Link href={home(lang)} className="wordmark srf">{t.siteName}</Link>
        <nav className="mnav">
          <span>{t.city}</span>
          <Suspense fallback={<span />}>
            <LangSwitch lang={lang} />
          </Suspense>
        </nav>
      </header>
      <div className="rule" />
      <main>{children}</main>
      <p className="foot">{t.footer}</p>
    </div>
  );
}
