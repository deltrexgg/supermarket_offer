// Package render turns HTML into images using headless Chrome via chromedp.
package render

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type Renderer struct {
	allocCtx context.Context
	cancel   context.CancelFunc
}

type Options struct {
	Width   int64
	Height  int64
	Scale   float64
	Format  string // "png" or "jpeg"
	Quality int    // jpeg only, 1-100
}

// New starts one long-lived headless Chrome process shared by all renders.
func New(chromePath string) (*Renderer, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true), // required inside containers
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	if chromePath != "" {
		opts = append(opts, chromedp.ExecPath(chromePath))
	}
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Warm up the browser so the first request is fast and startup errors surface early.
	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()
	warm, cancelWarm := context.WithTimeout(ctx, 30*time.Second)
	defer cancelWarm()
	if err := chromedp.Run(warm); err != nil {
		cancel()
		return nil, fmt.Errorf("start headless chrome: %w", err)
	}
	return &Renderer{allocCtx: allocCtx, cancel: cancel}, nil
}

func (r *Renderer) Close() { r.cancel() }

// HTMLToImage renders the given HTML document and returns the image bytes.
func (r *Renderer) HTMLToImage(ctx context.Context, html string, o Options) ([]byte, error) {
	if html == "" {
		return nil, errors.New("html is empty")
	}
	if o.Width <= 0 {
		o.Width = 1080
	}
	if o.Height <= 0 {
		o.Height = 1350
	}
	if o.Scale <= 0 {
		o.Scale = 1
	}
	if o.Format == "" {
		o.Format = "png"
	}

	tabCtx, cancelTab := chromedp.NewContext(r.allocCtx)
	defer cancelTab()
	tabCtx, cancelTimeout := context.WithTimeout(tabCtx, 30*time.Second)
	defer cancelTimeout()

	var buf []byte
	err := chromedp.Run(tabCtx,
		chromedp.EmulateViewport(o.Width, o.Height, chromedp.EmulateScale(o.Scale)),
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			tree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(tree.Frame.ID, html).Do(ctx)
		}),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			format := page.CaptureScreenshotFormatPng
			if o.Format == "jpeg" {
				format = page.CaptureScreenshotFormatJpeg
			}
			var err error
			buf, err = page.CaptureScreenshot().
				WithFormat(format).
				WithQuality(int64(o.Quality)).
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}
	return buf, nil
}
