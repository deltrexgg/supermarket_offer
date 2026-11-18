# ---- build ----
FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY . .
# go.sum is generated on first build (run `go mod tidy` locally and commit it to pin versions)
RUN go mod tidy && CGO_ENABLED=0 go build -o /out/server ./cmd/server

# ---- runtime (Chromium for HTML-to-image) ----
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
      chromium fonts-liberation fonts-noto-core fonts-noto-color-emoji \
      fonts-noto-ui-core ca-certificates curl \
    && rm -rf /var/lib/apt/lists/*

RUN useradd --create-home --uid 10001 app
USER app
WORKDIR /app
COPY --from=build /out/server /app/server

ENV CHROME_PATH=/usr/bin/chromium \
    GIN_MODE=release \
    PORT=8080
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --retries=5 \
  CMD curl -fs http://localhost:8080/health || exit 1
CMD ["/app/server"]
