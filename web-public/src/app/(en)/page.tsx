import type { Metadata } from "next";
import HomeView from "@/components/HomeView";
import { homeMetadata } from "@/lib/seo";

export const metadata: Metadata = homeMetadata("en");

export default function Home({ searchParams }: { searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  return <HomeView lang="en" searchParams={searchParams} />;
}
