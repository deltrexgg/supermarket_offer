import type { MetadataRoute } from "next";
import { connection } from "next/server";
import { getOffers } from "@/lib/api";
import { home, offerPath, SITE_URL } from "@/lib/i18n";

const MAX_PAGES = 50; // 50 pages of 48 offers is far more than a day's active offers

const url = (path: string) => `${SITE_URL}${path === "/" ? "" : path}`;

// Offers expire at Dubai midnight, so the list is kept for ten minutes and then rebuilt.
let cached: { at: number; ids: number[] } | null = null;

async function activeOfferIds(): Promise<number[]> {
  if (cached && Date.now() - cached.at < 10 * 60 * 1000) return cached.ids;
  const ids: number[] = [];
  for (let page = 1; page <= MAX_PAGES; page++) {
    const r = await getOffers(new URLSearchParams({ page: String(page), page_size: "48" }));
    ids.push(...r.items.map((o) => o.id));
    if (page * r.page_size >= r.total || r.items.length === 0) break;
  }
  cached = { at: Date.now(), ids };
  return ids;
}

// Built per request (not at image build time, when the API is not running yet).
export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  await connection();
  const pair = (enPath: string, arPath: string) => ({ languages: { en: url(enPath), ar: url(arPath) } });

  const entries: MetadataRoute.Sitemap = [
    { url: url(home("en")), changeFrequency: "daily", priority: 1, alternates: pair(home("en"), home("ar")) },
    { url: url(home("ar")), changeFrequency: "daily", priority: 1, alternates: pair(home("en"), home("ar")) },
  ];

  let ids: number[] = [];
  try {
    ids = await activeOfferIds();
  } catch (e) {
    // If the API is down, still serve the home pages instead of failing the crawler.
    console.error("sitemap: could not load offers", e);
  }
  for (const id of ids) {
    const alt = pair(offerPath("en", id), offerPath("ar", id));
    entries.push({ url: url(offerPath("en", id)), changeFrequency: "daily", priority: 0.8, alternates: alt });
    entries.push({ url: url(offerPath("ar", id)), changeFrequency: "daily", priority: 0.8, alternates: alt });
  }
  return entries;
}
