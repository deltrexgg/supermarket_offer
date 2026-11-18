# API contract (v1)

Base path: `/api/v1`. JSON in and out, except file uploads (`multipart/form-data`) and poster downloads (`image/png`).

## Conventions

- **Auth:** `Authorization: Bearer <JWT>`. The token carries `user_id`, `role` and `supermarket_id`. Passwords are hashed with bcrypt.
- **Roles:** `public` (no token), `supermarket`, `admin`. A route checks its role first, then ownership.
- **Tenant isolation:** the supermarket is always taken from the token, never from the URL or body. Another supermarket's record returns `404`, as if it did not exist.
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
| POST | `/auth/register` | public | Body: `name`, `email`, `password`, `phone`. Creates a supermarket (`pending`) and its user in one transaction. `201`. |
| POST | `/auth/login` | public | Body: `email`, `password`. Returns `{ token, user }`. `403` if the supermarket is blocked. |
| GET | `/auth/me` | any | Current user and supermarket status. |

## Public

Only offers that are active today in Dubai, from approved supermarkets.

| Method | Path | Notes |
|---|---|---|
| GET | `/offers` | Query: `q`, `area`, `category`, `supermarket`, `sort` (`lowest_price`, `highest_discount`, `ending_soon`), `page`, `page_size`. Cached in Redis. |
| GET | `/offers/{id}` | Offer detail with all branches (name, area, address, phone) and `days_left`. `404` if not active or not approved. |
| GET | `/areas` | Filter options. |
| GET | `/categories` | Filter options. |
| GET | `/supermarkets` | Approved supermarkets, for the filter. |

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
| POST | `/portal/offers/{id}/poster` | Body: `template` (`hot_deal`, `weekend_offer`), `lang` (`en`, `ar`). Renders the PNG and returns `{ poster_url }`. |
| GET | `/portal/offers/{id}/poster` | Downloads the PNG (`Content-Disposition: attachment`). |
| GET | `/portal/offers/{id}/poster/print` | Print-friendly A4 HTML view. |
| POST | `/portal/ai/offer-text` | Nice-to-have. Body: item details. Returns `description` and `item_name_ar` from Claude. |

Upload rules: JPEG, PNG or WebP, checked by file signature (not just extension), max 5 MB, stored under a random name.

## Admin (role `admin`)

| Method | Path | Notes |
|---|---|---|
| GET | `/admin/supermarkets` | `status` filter, paging. |
| PATCH | `/admin/supermarkets/{id}/status` | Body: `status` (`approved`, `blocked`, `pending`). Refreshes the public cache. |
| GET / POST | `/admin/areas`, `/admin/categories` | |
| PUT / DELETE | `/admin/areas/{id}`, `/admin/categories/{id}` | `409` if still in use. |
| GET | `/admin/dashboard` | `supermarkets_total`, `supermarkets_pending`, `offers_active_today`, `offers_expiring_today`. |

## Health

`GET /health` returns `{ "db": "ok", "redis": "ok" }` (`503` if either is down).
