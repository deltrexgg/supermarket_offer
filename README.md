# supermarket_offer

Offer platform for supermarkets: a merchant portal and a public deals website.

| Part | Tech | Port |
|---|---|---|
| `api/` | Go, Gin, GORM | 8080 |
| `web-portal/` | Next.js (merchant portal) | 3000 |
| `web-public/` | Next.js (public website) | 3001 |
| postgres | PostgreSQL 16 | 5432 |
| redis | Redis 7 | 6379 |

## Run

```bash
cp .env.example .env
docker compose up --build
```

- API health: http://localhost:8080/health
- Portal: http://localhost:3000
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
