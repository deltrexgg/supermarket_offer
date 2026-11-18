"use client";

import ErrorView from "@/components/ErrorView";

export default function Error({ error, retry }: { error: unknown; retry: () => void }) {
  return <ErrorView lang="ar" error={error} retry={retry} />;
}
