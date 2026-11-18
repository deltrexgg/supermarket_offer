"use client";

import { useEffect } from "react";
import { strings, type Lang } from "@/lib/i18n";

// Shown when a page throws while rendering, instead of a blank screen.
export default function ErrorView({ lang, error, retry }: { lang: Lang; error: unknown; retry: () => void }) {
  const t = strings(lang);
  useEffect(() => {
    console.error(error);
  }, [error]);
  return (
    <div className="empty" role="alert">
      <h1>{t.errorTitle}</h1>
      <p>{t.errorBody}</p>
      <button className="primary" onClick={() => retry()}>{t.retry}</button>
    </div>
  );
}
