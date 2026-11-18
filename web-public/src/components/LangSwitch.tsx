"use client";

import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";
import { strings, type Lang } from "@/lib/i18n";

// Links to the same page in the other language, keeping the search filters.
export default function LangSwitch({ lang }: { lang: Lang }) {
  const pathname = usePathname();
  const params = useSearchParams().toString();
  const search = params ? `?${params}` : "";

  const other: Lang = lang === "en" ? "ar" : "en";
  let target: string;
  if (lang === "en") target = pathname === "/" ? "/ar" : `/ar${pathname}`;
  else target = pathname === "/ar" ? "/" : pathname.replace(/^\/ar/, "") || "/";
  const t = strings(lang);

  return (
    <Link href={target + search} className="langswitch" hrefLang={other} lang={other} aria-label={t.langSwitchLabel}>
      {t.langSwitch}
    </Link>
  );
}
