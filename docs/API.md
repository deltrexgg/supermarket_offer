# API contract (v1)

Base path: `/api/v1`. JSON in and out, except file uploads (`multipart/form-data`) and poster downloads (`image/png`).

## Conventions

- **Auth:** Supabase Auth. The web apps sign in with `supabase-js` and send the Supabase access token as `Authorization: Bearer <token>`. The API verifies it against the project's public keys (JWKS), then looks the user up in `users` by `auth_user_id` to get the role and supermarket. Supabase stores the hashed passwords.
- **Roles:** `public` (no token), `supermarket`, `admin`. A route checks its role first, then ownership.
- **Tenant isolation:** the supermarket comes from the `users` row of the verified token, never from the URL or body. Another supermarket's record returns `404`, as if it did not exist.
- **Money:** AED, 2 decimals. **Dates:** `YYYY-MM-DD` Dubai calendar days.
- **Pagination:** `page` (default 1) and `page_size` (default 20, max 50). Responses include `total`.

### Error format

```json
{ "error": { "code": "validation_failed",
             "message": "Some fields are invalid.",
             "fields": { "offer_price": "must be lower than regular_price" } } }
```

| Status | When |
|---|---|
| 200 / 201 / 204 | OK / created / deleted |
| 400 | Malformed JSON or query |
| 401 | Missing or invalid token |
| 403 | Wrong role, or supermarket is pending or blocked for this action |
| 404 | Not found, or belongs to another supermarket |
| 409 | Conflict, such as duplicate email or branch name |
| 413 / 415 | Upload too large / unsupported file type |
| 422 | Validation failed (`fields` filled in) |
| 500 | Unexpected error (no internal details exposed) |

## Auth

| Method | Path | Role | Notes |
|---|---|---|---|
| POST | `/auth/register` | public | Body: `name`, `email`, `password`, `phone`. Creates the Supabase user (email pre-confirmed), then the supermarket (`pending`) and its `users` row in one transaction. If that transaction fails, the Supabase user is deleted again. `201`. `409` if the email exists. |
| GET | `/auth/me` | any signed-in | Current user, role and supermarket status. `403` if the supermarket is blocked. |

There is no `/auth/login` endpoint: signing in is done by the web app directly with Supabase Auth.

Public lookups used by the forms: `GET /areas` and `GET /categories` return `{ items: [{id, name}] }`.

Offers created or edited with `branch_ids[]` must list at least one of your own branches. A branch that offers still use cannot be deleted (`409`).

## Public

Only offers that are active today in Dubai, from approved supermarkets.

| Method | Path | Notes |
|---|---|---|
| GET | `/offers` | Query: `q`, `area`, `category`, `supermarket`, `sort` (`lowest_price`, `highest_discount`, `ending_soon`), `page`, `page_size`. Cached in Redis. |
| GET | `/offers/{id}` | Offer detail with all branches (name, area, address, phone) and `days_left`. `404` if not active or not approved. |
| GET | `/areas` | Filter options. |
| GET | `/categories` | Filter options. |
| GET | `/supermarkets` | Approved supermarkets, for the filter. |

Caching: list and detail responses are stored in Redis under a key made from the query, the Dubai date and a version counter. The counter is bumped after every offer or branch change in the portal and every admin change to a supermarket status, offer or lookup name, and entries expire at Dubai midnight. If Redis is down the API reads the database. The `X-Cache` response header is `HIT` or `MISS`.

Validation: invalid list filters (for example an unknown `sort`) return `422` with field errors.

An offer card contains: photo, item name, brand, unit, supermarket and logo, branch count, regular price, offer price, `discount_pct`, `days_left`.

## Supermarket portal (role `supermarket`)

| Method | Path | Notes |
|---|---|---|
| GET | `/portal/branches` | Own branches. |
| POST | `/portal/branches` | `name`, `area_id`, `address`, `phone`. `409` on duplicate name. |
| GET / PUT / DELETE | `/portal/branches/{id}` | |
| GET | `/portal/offers` | `status` filter (`upcoming`, `active`, `expired`), `q`, paging. |
| POST | `/portal/offers` | `multipart/form-data`: `item_name`, `brand`, `unit`, `category_id`, `regular_price`, `offer_price`, `valid_from`, `valid_to`, `branch_ids[]`, `photo`. `422` if any branch is not yours. Returns the offer with `discount_pct`. |
| GET / PUT / DELETE | `/portal/offers/{id}` | PUT may replace the photo and the branch list. Cache is refreshed after commit. |
| POST | `/portal/offers/{id}/poster` | Body: `template` (`hot_deal`, `weekend_offer`), `lang` (`en`, `ar`). Renders an A4 PNG (1240 x 1754) with headless Chrome, stores it as `posters/{offer_id}-{timestamp}.png` (a new name each time, so caches never show an old poster) and returns `{ poster_url }`. Editing the offer clears its poster. |
| GET | `/portal/offers/{id}/poster` | Downloads the PNG (`Content-Disposition: attachment`). |
| GET | `/portal/offers/{id}/poster/print` | Print-friendly A4 HTML view. Query: `template`, `lang`. It opens the print dialog; the portal fetches it with the login token and opens it in a new tab. |
| POST | `/portal/ai/offer-text` | Nice-to-have. Body: item details. Returns `description` and `item_name_ar` from Claude. |

Upload rules: JPEG, PNG, WebP, GIF or AVIF, checked by file signature (not just extension), max 5 MB, stored under a random name in the public Supabase Storage bucket `offer-media` (S3 API). Responses return a full file URL built from `PUBLIC_FILE_BASE_URL` and the stored key.

## Admin (role `admin`)

The admin login comes from `ADMIN_EMAIL` and `ADMIN_PASSWORD` in `.env`; the API creates it on startup.

| Method | Path | Notes |
|---|---|---|
| GET | `/admin/stats` | Dashboard counts: supermarkets by status, branches, offers active, upcoming, expired and ending within 2 days. |
| GET | `/admin/supermarkets` | `status` (`pending`, `approved`, `blocked`), `q`, paging. Includes login email, branch and offer counts. Pending first. |
| PATCH | `/admin/supermarkets/{id}/status` | Body: `status` (`approved`, `blocked`, `pending`). Refreshes the public cache. |
| GET / POST | `/admin/areas`, `/admin/categories` | List includes `in_use`. Names are 2 to 80 characters, unique ignoring case (`409`). |
| PUT / DELETE | `/admin/areas/{id}`, `/admin/categories/{id}` | Delete returns `409` while the area has branches or the category has offers. |
| GET | `/admin/offers` | Every supermarket's offers. `status`, `supermarket`, `q`, paging. |
| DELETE | `/admin/offers/{id}` | Removes the offer and its photo and poster files. Refreshes the public cache. |

## Health

`GET /health` returns `{ "db": "ok", "redis": "ok" }` (`503` if either is down).
