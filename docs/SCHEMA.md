# Database schema

PostgreSQL 16. The schema is created by `api/migrations/0001_init.sql`, applied automatically when the API starts. GORM never creates or alters tables.

## Tables

| Table | Purpose | Key columns |
|---|---|---|
| `areas` | Dubai areas used for branches and the public filter (admin-managed) | `name` (unique, case-insensitive) |
| `categories` | Offer categories (admin-managed) | `name` (unique, case-insensitive) |
| `supermarkets` | The tenants | `slug` (unique), `status`: `pending`, `approved`, `blocked` |
| `users` | Logins for admins and supermarkets | `email` (unique, case-insensitive), `password_hash`, `role`, `supermarket_id` (unique) |
| `branches` | A supermarket's shops | `supermarket_id`, `area_id`, unique `(supermarket_id, name)` |
| `offers` | One item offer | `item_name`, `brand`, `unit`, `regular_price`, `offer_price`, `discount_pct`, `valid_from`, `valid_to`, `photo_path`, `poster_path` |
| `offer_branches` | Which branches an offer applies to | `(offer_id, branch_id)`, `supermarket_id` |
| `schema_migrations` | Applied migrations with checksums | `version`, `checksum` |

View `offers_with_status` adds `status` (`upcoming`, `active`, `expired`) and `days_left` (0 means "Ends today") to every offer.

## Rules the database enforces

- **Prices:** both greater than 0, and `offer_price < regular_price`.
- **Discount:** `discount_pct` is a generated column, rounded to a whole percent. It is read-only, so it can never disagree with the prices.
- **Dates:** `valid_to >= valid_from`. They are plain dates, meaning Dubai calendar days, inclusive of both ends.
- **Roles:** an `admin` user has no supermarket; a `supermarket` user must have one. Each supermarket has one login.
- **Tenant safety:** `offer_branches` uses composite foreign keys `(offer_id, supermarket_id)` and `(branch_id, supermarket_id)`. An offer can never be linked to another supermarket's branch, even if application code is wrong.
- **In-use data:** an area or category that is still referenced cannot be deleted (`RESTRICT`).
- **Cleanup:** deleting a supermarket removes its users, branches and offers. Deleting an offer or branch removes its `offer_branches` rows.
- **Timestamps:** `updated_at` is maintained by a trigger.

## Status is derived

Status is not stored. The view compares today's date in `Asia/Dubai` with `valid_from` and `valid_to`, so it can never go stale and expired offers cannot appear by mistake.

## ACID notes

- **Atomicity:** multi-step writes (register supermarket plus user, create or edit an offer plus its branch links, approve or block) run in one transaction.
- **Consistency:** the constraints above.
- **Isolation:** default `READ COMMITTED`. Editing an offer first locks it with `SELECT ... FOR UPDATE`. Duplicate sign-ups are settled by the unique constraints.
- **Durability:** default WAL and `synchronous_commit`, data kept in the `pgdata` volume. Redis is a cache only and never the source of truth.
- **Files:** poster and photo files cannot join a transaction. The offer is committed first, then the poster is rendered and `poster_path` is updated. A crash in between leaves a poster that can be regenerated, never a corrupt offer.

## Seed data

A seed (3 supermarkets, 6 branches in 4 areas, 30 offers) will use dates relative to today in Dubai, so the mix of upcoming, active and expired offers is the same on any day.
