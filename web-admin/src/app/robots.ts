import type { MetadataRoute } from "next";

// This site is for signed-in users only: keep it out of search results.
export default function robots(): MetadataRoute.Robots {
  return { rules: [{ userAgent: "*", disallow: "/" }] };
}
