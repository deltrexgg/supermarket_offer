export type Lang = "en" | "ar";

export const SITE_URL = (process.env.SITE_URL ?? "http://localhost:3001").replace(/\/$/, "");

// English lives at the root (/, /offers/5) and Arabic under /ar (/ar, /ar/offers/5).
export const home = (lang: Lang) => (lang === "ar" ? "/ar" : "/");
export const offerPath = (lang: Lang, id: number | string) => (lang === "ar" ? `/ar/offers/${id}` : `/offers/${id}`);

const en = {
  siteName: "Dubai Offers",
  heroTitle: "Today's supermarket offers in Dubai",
  heroSub: "Compare live deals from supermarkets near you. Prices in AED.",
  metaTitle: "Dubai Supermarket Offers",
  metaDesc: "Today's supermarket offers in Dubai: compare prices, discounts and branches.",
  loadingOffers: "Loading offers…",
  loadingOffer: "Loading offer…",
  unavailable: "Offers are unavailable right now. Please try again in a moment.",
  offerUnavailable: "This offer can't be loaded right now. Please try again in a moment.",
  searchPlaceholder: "Search items or brands",
  searchLabel: "Search offers",
  search: "Search",
  area: "Area",
  category: "Category",
  supermarket: "Supermarket",
  sortBy: "Sort by",
  all: "All",
  highestDiscount: "Highest discount",
  lowestPrice: "Lowest price",
  endingSoon: "Ending soon",
  clearFilters: "Clear filters",
  emptyBefore: "No offers match your search. Try a different word, or ",
  emptyLink: "clear the filters",
  emptyAfter: ".",
  previous: "Previous",
  next: "Next",
  pagesLabel: "Pages",
  back: "All offers",
  availableAt: "Available at",
  footer: "Prices in AED. Offers are valid on the dates shown, Dubai time.",
  notFoundTitle: "This offer isn't available",
  notFoundBody: "It may have ended, or it isn't published yet.",
  notFoundLink: "See today's offers",
  pageNotFound: "Page not found",
  pageNotFoundBody: "That address does not exist.",
  errorTitle: "Something went wrong",
  errorBody: "We could not load this page. Please try again in a moment.",
  retry: "Try again",
  langSwitch: "العربية",
  langSwitchLabel: "Switch to Arabic",
  was: "was",
  topOffer: "Top offer",
  viewOffer: "View offer",
  carouselLabel: "Top offers",
  pauseSlides: "Pause auto-slide",
  playSlides: "Play auto-slide",
  prevSlide: "Previous offer",
  nextSlide: "Next offer",
  catTitle: "Shop by category",
  offersTitle: "Today's offers",
  filters: "Filters",
  apply: "Apply filters",
  branchesTitle: "Available at",
  call: "Call",
};

const ar: typeof en = {
  siteName: "عروض دبي",
  heroTitle: "عروض السوبرماركت اليوم في دبي",
  heroSub: "قارن العروض الحية من السوبرماركت القريبة منك. الأسعار بالدرهم الإماراتي.",
  metaTitle: "عروض السوبرماركت في دبي",
  metaDesc: "عروض السوبرماركت اليوم في دبي: قارن الأسعار والخصومات والفروع.",
  loadingOffers: "جارٍ تحميل العروض…",
  loadingOffer: "جارٍ تحميل العرض…",
  unavailable: "العروض غير متاحة حالياً. يرجى المحاولة بعد قليل.",
  offerUnavailable: "تعذّر تحميل هذا العرض حالياً. يرجى المحاولة بعد قليل.",
  searchPlaceholder: "ابحث عن منتج أو علامة تجارية",
  searchLabel: "البحث في العروض",
  search: "بحث",
  area: "المنطقة",
  category: "الفئة",
  supermarket: "السوبرماركت",
  sortBy: "ترتيب حسب",
  all: "الكل",
  highestDiscount: "أعلى خصم",
  lowestPrice: "أقل سعر",
  endingSoon: "تنتهي قريباً",
  clearFilters: "مسح الفلاتر",
  emptyBefore: "لا توجد عروض مطابقة لبحثك. جرّب كلمة أخرى أو ",
  emptyLink: "امسح الفلاتر",
  emptyAfter: ".",
  previous: "السابق",
  next: "التالي",
  pagesLabel: "الصفحات",
  back: "كل العروض",
  availableAt: "متوفر في",
  footer: "الأسعار بالدرهم الإماراتي. العروض سارية في التواريخ المعروضة بتوقيت دبي.",
  notFoundTitle: "هذا العرض غير متاح",
  notFoundBody: "ربما انتهى العرض أو لم يُنشر بعد.",
  notFoundLink: "شاهد عروض اليوم",
  pageNotFound: "الصفحة غير موجودة",
  pageNotFoundBody: "هذا العنوان غير موجود.",
  errorTitle: "حدث خطأ ما",
  errorBody: "تعذّر تحميل هذه الصفحة. يرجى المحاولة بعد قليل.",
  retry: "حاول مرة أخرى",
  langSwitch: "English",
  langSwitchLabel: "التبديل إلى الإنجليزية",
  was: "كان",
  topOffer: "أفضل عرض",
  viewOffer: "عرض التفاصيل",
  carouselLabel: "أفضل العروض",
  pauseSlides: "إيقاف التمرير التلقائي",
  playSlides: "تشغيل التمرير التلقائي",
  prevSlide: "العرض السابق",
  nextSlide: "العرض التالي",
  catTitle: "تسوق حسب الفئة",
  offersTitle: "عروض اليوم",
  filters: "تصفية",
  apply: "تطبيق التصفية",
  branchesTitle: "متوفر في",
  call: "اتصال",
};

export const dict = { en, ar };
export const strings = (lang: Lang) => dict[lang];

// "12 offers" / "12 عرضاً" with Arabic number agreement.
export function offersCount(lang: Lang, n: number): string {
  if (lang === "en") return `${n} ${n === 1 ? "offer" : "offers"}`;
  if (n === 0) return "لا توجد عروض";
  if (n === 1) return "عرض واحد";
  if (n === 2) return "عرضان";
  return n <= 10 ? `${n} عروض` : `${n} عرضاً`;
}

export function branchesCount(lang: Lang, n: number): string {
  if (lang === "en") return `${n} ${n === 1 ? "branch" : "branches"}`;
  if (n === 1) return "فرع واحد";
  if (n === 2) return "فرعان";
  return n <= 10 ? `${n} فروع` : `${n} فرعاً`;
}

export function daysLeftLabel(lang: Lang, days: number): string {
  if (lang === "en") return days <= 0 ? "Ends today" : days === 1 ? "1 day left" : `${days} days left`;
  if (days <= 0) return "ينتهي اليوم";
  if (days === 1) return "باقي يوم واحد";
  if (days === 2) return "باقي يومان";
  return days <= 10 ? `باقي ${days} أيام` : `باقي ${days} يوماً`;
}

export function pageLabel(lang: Lang, page: number, pages: number): string {
  return lang === "en" ? `Page ${page} of ${pages}` : `صفحة ${page} من ${pages}`;
}

export function priceLabel(lang: Lang, n: number): string {
  return lang === "en" ? `AED ${n.toFixed(2)}` : `${n.toFixed(2)} د.إ`;
}

const MONTHS_EN = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
const MONTHS_AR = ["يناير", "فبراير", "مارس", "أبريل", "مايو", "يونيو", "يوليو", "أغسطس", "سبتمبر", "أكتوبر", "نوفمبر", "ديسمبر"];

export function longDate(lang: Lang, iso: string): string {
  const [y, m, d] = iso.split("-").map(Number);
  return `${d} ${(lang === "ar" ? MONTHS_AR : MONTHS_EN)[m - 1]} ${y}`;
}

export function validRange(lang: Lang, from: string, to: string): string {
  return lang === "en"
    ? `Valid ${longDate(lang, from)} to ${longDate(lang, to)}`
    : `صالح من ${longDate(lang, from)} إلى ${longDate(lang, to)}`;
}

// The shopper's language falls back to English when there is no Arabic text.
export const pick = (lang: Lang, en: string | null | undefined, ar: string | null | undefined): string =>
  (lang === "ar" && ar && ar.trim() ? ar : en) ?? "";

// Colour class for the "how soon" chip: red up to 2 days, amber up to 5, green after.
export function urgency(days: number): "bad" | "warn" | "ok" {
  return days <= 2 ? "bad" : days <= 5 ? "warn" : "ok";
}

export const offLabel = (lang: Lang, pct: number) => (lang === "en" ? `${pct}% OFF` : `خصم ${pct}%`);

export const saveLabel = (lang: Lang, regular: number, offer: number) =>
  lang === "en" ? `Save ${priceLabel(lang, regular - offer)}` : `وفّر ${priceLabel(lang, regular - offer)}`;

export const goToSlide = (lang: Lang, n: number) => (lang === "en" ? `Go to offer ${n}` : `الانتقال إلى العرض ${n}`);
