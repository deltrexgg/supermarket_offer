"use client";

import { useEffect } from "react";

// Shown when a page throws while rendering, instead of a blank screen.
export default function Error({ error, retry }: { error: unknown; retry: () => void }) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="empty" role="alert">
      <h1>Something went wrong</h1>
      <p>We could not load this page. Please try again in a moment.</p>
      <button className="primary" onClick={() => retry()}>Try again</button>
    </div>
  );
}
