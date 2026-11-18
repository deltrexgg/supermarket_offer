module github.com/deltrexgg/supermarket_offer/api

go 1.24.7

// Pinned: newer chromedp releases changed chromedp.Run to a generic API that
// internal/render does not use. go mod tidy keeps this version.
require github.com/chromedp/chromedp v0.14.2
