"use client";

import { useEffect } from "react";

// Shown when a page throws while rendering, instead of a blank screen.
export default function Error({ error, retry }: { error: unknown; retry: () => void }) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <main className="auth-card" role="alert">
      <h1>Something went wrong</h1>
      <p className="muted">The page could not be shown. Check that the API is running, then try again.</p>
      <button className="primary" onClick={() => retry()}>Try again</button>
    </main>
  );
}
