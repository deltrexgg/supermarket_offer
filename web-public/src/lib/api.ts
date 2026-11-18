// On the server (page rendering inside Docker) the API is reached by its service
// name; the browser uses the public address. API_INTERNAL_URL is only set on the server.
const BASE =
  (typeof window === "undefined" ? process.env.API_INTERNAL_URL : undefined) ??
  process.env.NEXT_PUBLIC_API_URL ??
  "http://localhost:8080";

export type Card = {
  id: number; item_name: string; item_name_ar: string | null; brand: string | null; unit: string;
  regular_price: number; offer_price: number; discount_pct: number; valid_to: string; days_left: number;
  photo_url: string | null; supermarket_id: number; supermarket_name: string; logo_url: string | null;
  branch_count: number;
};
export type Branch = { id: number; name: string; address: string; phone: string | null; area_id: number; area_name: string };
export type Detail = Card & {
  valid_from: string; description: string | null; description_ar: string | null;
  category_id: number; category_name: string; supermarket_phone: string | null; branches: Branch[];
};
export type OfferList = { items: Card[]; total: number; page: number; page_size: number };
export type Option = { id: number; name: string };

export class NotFoundError extends Error {}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}/api/v1${path}`, { cache: "no-store" });
  if (res.status === 404) throw new NotFoundError();
  if (!res.ok) throw new Error(`API ${path} failed (${res.status})`);
  return res.json() as Promise<T>;
}

export const getOffers = (qs: URLSearchParams) => get<OfferList>(`/offers?${qs}`);
export const getOffer = (id: string) => get<Detail>(`/offers/${encodeURIComponent(id)}`);
export const getAreas = () => get<{ items: Option[] }>("/areas").then((r) => r.items);
export const getCategories = () => get<{ items: Option[] }>("/categories").then((r) => r.items);
export const getSupermarkets = () => get<{ items: Option[] }>("/supermarkets").then((r) => r.items);
