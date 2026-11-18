# supermarket_offer

Offer platform for supermarkets: a merchant portal and a public deals website.

| Part | Tech | Port |
|---|---|---|
| `api/` | Go, Gin, GORM | 8080 |
| `web-portal/` | Next.js (merchant portal and admin) | 3000 |
| `web-public/` | Next.js (public website) | 3001 |
| Database | PostgreSQL on Supabase | hosted |
| Login | Supabase Auth | hosted |
| Images | Supabase Storage (S3 bucket) | hosted |
| redis | Redis 7 (docker-compose) | 6379 |

Design docs are in `docs/` (`SCHEMA.md`, `API.md`, `MODULES.md`).

## Set up Supabase (once)

1. Create a Supabase project.
2. **Database:** Project Settings > Database > Connection string > **Session pooler**. Copy it into `DATABASE_URL` (port 5432, `sslmode=require`). Do not use the transaction pooler (port 6543): the migration runner needs a session-level lock and GORM uses prepared statements.
3. **Auth:** Project Settings > API. Copy the project URL, the public (anon) key and the service-role key. The service-role key is secret and only the API uses it.
4. **Storage:** create a **public** bucket named `offer-media`, then Storage > S3 Connection to get the endpoint, region and access keys.
5. `cp .env.example .env` and fill in the values. Never commit `.env`.

## Run

```bash
docker compose up --build
```

On start the API applies the SQL migrations in `api/migrations/` to your Supabase database (look for `migrate: applied 0001_init.sql` in the log).

With `SEED=true` (the default in `.env.example`) the API also fills an empty database with Dubai demo data (3 supermarkets, 6 branches, 30 offers with dates relative to today) and creates these logins in your Supabase Auth. Password for all: `Offers@12345`.

| Login | Role |
|---|---|
| `admin@offers.test` | admin |
| `noor@offers.test` | Al Noor Fresh Mart |
| `gulf@offers.test` | Gulf Basket |
| `spice@offers.test` | Spice Garden Hypermarket |

Set `SEED=false` afterwards. The seed never touches a database that already has supermarkets.

- API health: http://localhost:8080/health
- Portal: http://localhost:3000 (sign in with a demo login, or register a new supermarket; branches, offers and photo upload)
- Public site: http://localhost:3001

## HTML to image

The API image bundles Chromium. `chromedp` drives it headless to render HTML as an image:

```bash
curl -X POST http://localhost:8080/api/v1/render \
  -H 'Content-Type: application/json' \
  -d '{"html":"<h1 style=\"font-family:sans-serif\">Fresh Milk 20% off</h1>","width":800,"height":400}' \
  --output flyer.png
```

Options: `width`, `height`, `scale` (max 3), `format` (`png` or `jpeg`), `quality` (jpeg).

## Notes

- `api/go.sum` is generated on the first Docker build. Run `go mod tidy` in `api/` and commit it to pin versions.
- The `/render` endpoint is unauthenticated for now. Put it behind auth before exposing it publicly.
- Login code (Supabase token verification) and image uploads to the bucket are built in later phases. Their settings are already in `.env.example`.
