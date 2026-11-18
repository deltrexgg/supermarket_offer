import { supabase } from "./supabase";

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  status: number;
  code: string;
  fields: Record<string, string>;
  constructor(status: number, code: string, message: string, fields: Record<string, string> = {}) {
    super(message);
    this.status = status;
    this.code = code;
    this.fields = fields;
  }
}

// Calls the Go API with the signed-in user's Supabase access token.
// Pass FormData for uploads (the browser sets the multipart header itself).
export async function api<T>(path: string, init: { method?: string; body?: unknown; auth?: boolean } = {}): Promise<T> {
  const headers: Record<string, string> = {};
  if (init.auth !== false) {
    const { data } = await supabase().auth.getSession();
    if (data.session) headers.Authorization = `Bearer ${data.session.access_token}`;
  }
  let body: BodyInit | undefined;
  if (init.body instanceof FormData) {
    body = init.body;
  } else if (init.body !== undefined) {
    headers["Content-Type"] = "application/json";
    body = JSON.stringify(init.body);
  }

  let res: Response;
  try {
    res = await fetch(`${API}/api/v1${path}`, { method: init.method ?? "GET", headers, body });
  } catch {
    throw new ApiError(0, "network", "Cannot reach the server. Is the API running?");
  }
  if (res.status === 204) return undefined as T;

  const json = await res.json().catch(() => null);
  if (!res.ok) {
    const e = json?.error;
    throw new ApiError(res.status, e?.code ?? "error", e?.message ?? `Request failed (${res.status})`, e?.fields ?? {});
  }
  return json as T;
}

export type Me = { id: number; email: string; role: "admin" | "supermarket"; supermarket_id: number | null };
export type Area = { id: number; name: string };
export type Category = { id: number; name: string };
export type Branch = {
  id: number; name: string; address: string; phone: string | null; area_id: number; area_name: string;
};
export type Offer = {
  id: number; item_name: string; item_name_ar: string | null; brand: string | null; unit: string;
  category_id: number; regular_price: number; offer_price: number; discount_pct: number;
  valid_from: string; valid_to: string; status: "upcoming" | "active" | "expired"; days_left: number;
  photo_url: string | null; poster_url: string | null; branch_ids: number[];
};
export type Page<T> = { items: T[]; total: number; page: number; page_size: number };
