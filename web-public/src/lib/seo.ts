import type { Metadata } from "next";
import { getOffer } from "@/lib/api";
import { home, offerPath, pick, priceLabel, SITE_URL, strings, type Lang } from "@/lib/i18n";

const LOCALE: Record<Lang, string> = { en: "en_AE", ar: "ar_AE" };

// Links between the English and Arabic version of a page, for search engines.
function alternates(enPath: string, arPath: string, lang: Lang): Metadata["alternates"] {
  return {
    canonical: lang === "ar" ? arPath : enPath,
    languages: { en: enPath, ar: arPath, "x-default": enPath },
  };
}

export function homeMetadata(lang: Lang): Metadata {
  const t = strings(lang);
  return {
    title: { absolute: t.metaTitle },
    description: t.metaDesc,
    alternates: alternates(home("en"), home("ar"), lang),
    openGraph: { title: t.metaTitle, description: t.metaDesc, url: home(lang), locale: LOCALE[lang], siteName: t.siteName, type: "website" },
  };
}

// Title and description for one offer. Falls back to the site title if the API is unreachable.
export async function offerMetadata(lang: Lang, id: string): Promise<Metadata> {
  try {
    const o = await getOffer(id);
    const name = pick(lang, o.item_name, o.item_name_ar);
    const off = o.discount_pct > 0 ? (lang === "ar" ? ` خصم ${o.discount_pct}%` : ` ${o.discount_pct}% off`) : "";
    const title = lang === "ar" ? `${name}${off} في ${o.supermarket_name}` : `${name}${off} at ${o.supermarket_name}`;
    const description =
      pick(lang, o.description, o.description_ar) ||
      (lang === "ar"
        ? `${name} بسعر ${priceLabel(lang, o.offer_price)} بدلاً من ${priceLabel(lang, o.regular_price)}.`
        : `${name} for ${priceLabel(lang, o.offer_price)} instead of ${priceLabel(lang, o.regular_price)}.`);
    const images = o.photo_url ? [o.photo_url] : undefined;
    return {
      title,
      description,
      alternates: alternates(offerPath("en", o.id), offerPath("ar", o.id), lang),
      openGraph: { title, description, url: offerPath(lang, o.id), locale: LOCALE[lang], type: "website", images },
    };
  } catch {
    return { title: strings(lang).metaTitle };
  }
}

export const metadataBase = new URL(SITE_URL);
