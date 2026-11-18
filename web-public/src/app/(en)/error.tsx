"use client";

import ErrorView from "@/components/ErrorView";

export default function Error({ error, retry }: { error: unknown; retry: () => void }) {
  return <ErrorView lang="en" error={error} retry={retry} />;
}
