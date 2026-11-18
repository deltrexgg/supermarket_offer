# Module plan

## Services

| Part | Folder | Port |
|---|---|---|
| API | `api/` (Go, Gin, GORM) | 8080 |
| Supermarket portal | `web-portal/` (Next.js) | 3000 |
| Platform admin | `web-admin/` (Next.js) | 3002 |
| Public offers website | `web-public/` (Next.js) | 3001 |
| PostgreSQL, login (Supabase Auth), image storage (S3 bucket) | Supabase (hosted, not in docker-compose) | n/a |
| Redis 7 | docker-compose | 6379 |

## API packages (`api/internal/`)

| Package | Responsibility | Phase |
|---|---|---|
| `config`, `db`, `cache`, `migrate` | Settings, Postgres, Redis, SQL migrations (done) | scaffold |
| `models` | GORM structs matching the tables (done) | scaffold |
| `dubai` | `Today()`, `DateOffset`, date parsing and time until midnight in `Asia/Dubai` | 1 |
| `auth` | Verifies Supabase access tokens (JWKS), looks up the `users` row, `RequireRole` middleware, Supabase admin client for register | 1 |
| `tenant` | The only way to query tenant data: `Scope(supermarketID)` | 1 |
| `httpapi` | Router, CORS, security headers, body limits | 1 |
| `apierr` | The single error format | 1 |
| `storage` | `Storage` interface with an S3 implementation for Supabase Storage (bucket `offer-media`) | 2 |
| `portal` | Branch and offer handlers and services | 2 |
| `poster`, `render` | HTML templates and the Chromium renderer | 3 |
| `public` | Listing, search, filters, sorting, Redis cache | 4 |
| `admin` | Dashboard stats, supermarket approval, areas, categories, offer removal; `seed.EnsureAdmin` creates the env admin login | 5 |
| `ratelimit` | Per-IP fixed-window limits counted in Redis | 6 |
| `ai` | Claude API for descriptions and Arabic names (nice-to-have) | 7 |

## Key flows

### Tenant isolation
1. The user signs in with Supabase Auth. The API verifies the access token, then loads the matching `users` row (by `auth_user_id`), which holds the role and `supermarket_id`. A login with no `users` row is refused. (A blocked supermarket can still sign in to its portal; it is only hidden from the public site.)
2. Middleware puts the supermarket ID in the request context. Handlers never read it from the URL or body.
3. Every portal query goes through `tenant.Scope`, which adds `WHERE supermarket_id = ?`. A record from another supermarket is simply not found, so the API answers `404`.
4. When an offer is saved, each submitted `branch_id` is checked against the same scope.
5. Second layer: the composite foreign keys in the database reject cross-tenant links even if a check is missed.

### Dubai time
- `dubaitime.Today()` is the only place that decides "today". The server and browser time zones are never used.
- The DB view `offers_with_status` uses the same rule, so the portal tabs and the public site always agree.
- Go embeds timezone data (`time/tzdata`), because the slim runtime image has none.
- The countdown in the browser is computed from `days_left` sent by the API, not from the browser clock.

### Redis cache (public listing and search)
- Key: `offers:v{version}:{dubai-date}:{hash of query}`. TTL: until the next Dubai midnight, so expired offers disappear without a cleanup job.
- Any change to offers, branches, supermarket status, or an area or category name runs `INCR offers:version` **after** the database commit. Old keys are never read again and expire on their own.
- If Redis is down, the API serves from Postgres (slower, still correct).

### Poster generation
1. The offer is committed to the database.
2. A Go HTML template (auto-escaped) renders the chosen template and language, using the offer, supermarket, logo and photo.
3. The existing `chromedp` renderer turns it into an A4 PNG (1240 x 1754).
4. The PNG is saved through `storage`, and `poster_path` is updated.
5. The print view serves the same HTML with A4 print CSS.
6. Regenerated when the offer is edited. Fonts for English and Arabic are installed in the API image.

### Uploads
Allowed types are checked by file signature, max 5 MB, stored under random names as `offers/{supermarket_id}/{uuid}.{ext}` (posters as `posters/{offer_id}-{timestamp}.png`, a new name each time so caches never show an old poster). Allowed: JPEG, PNG, WebP, GIF, AVIF in the public Supabase bucket `offer-media`. The database stores only the key. The public URL is `PUBLIC_FILE_BASE_URL` + key.

### Supabase credentials
The API holds the service-role key and the S3 keys. The portal gets only the project URL and the public (anon) key, and the public site gets none. Nothing secret is sent to a browser or committed to git.

## Frontend

**`web-portal`** (client-rendered, behind login): `/login`, `/register`, `/branches`, `/offers` (list, form and poster panel on one page).

**`web-admin`** (client-rendered, admin role only): `/login`, `/dashboard`, `/supermarkets`, `/lists` (areas and categories), `/offers`.

**`web-public`** (server-rendered): `/` (listing, search, filters, paging), `/offers/[id]` (detail with branches). English only for now; Arabic with right-to-left layout is planned for phase 7.

## Delivery phases

0 plan and design, 1 foundation, 2 supermarket portal, 3 poster engine, 4 public site and cache, 5 admin, 6 hardening and docs, 7 nice-to-haves (Claude AI and Arabic).
