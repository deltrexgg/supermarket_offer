import Link from "next/link";
import { Suspense } from "react";
import Icon from "@/components/Icon";
import LangSwitch from "@/components/LangSwitch";
import { home, strings, type Lang } from "@/lib/i18n";

// Header, content and footer shared by the English and Arabic sites.
export default function Shell({ lang, children }: { lang: Lang; children: React.ReactNode }) {
  const t = strings(lang);
  return (
    <>
      <header className="top">
        <div className="wrap bar-top">
          <Link href={home(lang)} className="logo"><i><Icon name="tag" size={24} /></i> {t.siteName}</Link>
          <Suspense fallback={<span />}>
            <LangSwitch lang={lang} />
          </Suspense>
        </div>
      </header>
      <main className="wrap">{children}</main>
      <footer className="wrap muted foot">{t.footer}</footer>
    </>
  );
}
