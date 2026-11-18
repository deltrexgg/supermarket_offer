package httpapi

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/deltrexgg/supermarket_offer/api/internal/admin"
	"github.com/deltrexgg/supermarket_offer/api/internal/apierr"
	"github.com/deltrexgg/supermarket_offer/api/internal/auth"
	"github.com/deltrexgg/supermarket_offer/api/internal/portal"
	"github.com/deltrexgg/supermarket_offer/api/internal/public"
	"github.com/deltrexgg/supermarket_offer/api/internal/ratelimit"
	"github.com/deltrexgg/supermarket_offer/api/internal/render"
)

type Deps struct {
	DB          *gorm.DB
	Redis       *redis.Client
	Renderer    *render.Renderer
	Auth        *auth.Verifier
	Portal      *portal.Handler
	Public      *public.Handler
	Admin       *admin.Handler
	CORSOrigins string

	// TrustedProxies: proxies whose X-Forwarded-For is believed (empty: none).
	TrustedProxies []string
	// RateLimitPublic: requests per minute per IP on public routes (0 turns limits off).
	RateLimitPublic int
}

func NewRouter(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	// Without this gin believes X-Forwarded-For from anyone, which would let a client
	// choose its own IP and dodge the rate limits.
	if err := r.SetTrustedProxies(d.TrustedProxies); err != nil {
		log.Fatalf("config: TRUSTED_PROXIES: %v", err)
	}
	r.Use(securityHeaders())
	counter := ratelimit.NewRedisCounter(d.Redis)
	limit := func(name string, n int, window time.Duration) gin.HandlerFunc {
		return ratelimit.Middleware(counter, name, n, window)
	}
	perMinute := d.RateLimitPublic
	r.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(d.CORSOrigins, ","),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		status := gin.H{"db": "ok", "redis": "ok"}
		code := http.StatusOK

		if sqlDB, err := d.DB.DB(); err != nil || sqlDB.PingContext(ctx) != nil {
			status["db"] = "down"
			code = http.StatusServiceUnavailable
		}
		if err := d.Redis.Ping(ctx).Err(); err != nil {
			status["redis"] = "down"
			code = http.StatusServiceUnavailable
		}
		c.JSON(code, status)
	})

	r.NoRoute(func(c *gin.Context) { apierr.NotFound(c, "route not found") })

	v1 := r.Group("/api/v1")

	// Public routes.
	v1.POST("/auth/register", limit("register", 10, time.Hour), bodyLimit(64<<10), d.Portal.Register)
	pub := v1.Group("", limit("public", perMinute, time.Minute))
	pub.GET("/areas", d.Portal.Areas)
	pub.GET("/categories", d.Portal.Categories)
	pub.GET("/offers", d.Public.ListOffers)
	pub.GET("/offers/:id", d.Public.GetOffer)
	pub.GET("/supermarkets", d.Public.ListSupermarkets)

	// Signed-in routes. Phase 2+ adds the supermarket and admin groups here.
	authed := v1.Group("", limit("authed", 300, time.Minute), noStore(), d.Auth.Authenticate())
	authed.GET("/auth/me", func(c *gin.Context) { c.JSON(http.StatusOK, auth.CurrentUser(c)) })

	// Supermarket portal: signed in AND role supermarket. The supermarket comes from
	// the verified user, so nothing in the URL or body can point at another tenant.
	pt := authed.Group("/portal", auth.RequireRole("supermarket"))
	pt.GET("/branches", d.Portal.ListBranches)
	pt.POST("/branches", bodyLimit(64<<10), d.Portal.CreateBranch)
	pt.GET("/branches/:id", d.Portal.GetBranch)
	pt.PUT("/branches/:id", bodyLimit(64<<10), d.Portal.UpdateBranch)
	pt.DELETE("/branches/:id", d.Portal.DeleteBranch)
	pt.GET("/offers", d.Portal.ListOffers)
	pt.POST("/offers", bodyLimit(6<<20+64<<10), d.Portal.CreateOffer)
	pt.GET("/offers/:id", d.Portal.GetOffer)
	pt.PUT("/offers/:id", bodyLimit(6<<20+64<<10), d.Portal.UpdateOffer)
	pt.DELETE("/offers/:id", d.Portal.DeleteOffer)
	pt.POST("/ai/offer-text", limit("ai", 10, time.Minute), bodyLimit(8<<10), d.Portal.AIOfferText)
	pt.POST("/offers/:id/poster", limit("poster", 20, time.Minute), bodyLimit(4<<10), d.Portal.CreatePoster)
	pt.GET("/offers/:id/poster", d.Portal.DownloadPoster)
	pt.GET("/offers/:id/poster/print", d.Portal.PrintPoster)

	// Platform owner: signed in AND role admin.
	ad := authed.Group("/admin", limit("admin", 120, time.Minute), auth.RequireRole("admin"))
	ad.GET("/stats", d.Admin.Stats)
	ad.GET("/supermarkets", d.Admin.ListSupermarkets)
	ad.PATCH("/supermarkets/:id/status", bodyLimit(4<<10), d.Admin.SetSupermarketStatus)
	ad.GET("/areas", d.Admin.ListAreas)
	ad.POST("/areas", bodyLimit(4<<10), d.Admin.CreateArea)
	ad.PUT("/areas/:id", bodyLimit(4<<10), d.Admin.RenameArea)
	ad.DELETE("/areas/:id", d.Admin.DeleteArea)
	ad.GET("/categories", d.Admin.ListCategories)
	ad.POST("/categories", bodyLimit(4<<10), d.Admin.CreateCategory)
	ad.PUT("/categories/:id", bodyLimit(4<<10), d.Admin.RenameCategory)
	ad.DELETE("/categories/:id", d.Admin.DeleteCategory)
	ad.GET("/offers", d.Admin.ListOffers)
	ad.DELETE("/offers/:id", d.Admin.DeleteOffer)

	// Rendering runs headless Chrome, so only signed-in users may call it.
	authed.POST("/render", limit("render", 10, time.Minute), renderHandler(d.Renderer))

	return r
}

type renderRequest struct {
	HTML    string  `json:"html" binding:"required"`
	Width   int64   `json:"width"`
	Height  int64   `json:"height"`
	Scale   float64 `json:"scale"`
	Format  string  `json:"format"` // png | jpeg
	Quality int     `json:"quality"`
}

// renderHandler converts HTML to an image with headless Chrome.
func renderHandler(rn *render.Renderer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req renderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apierr.BadRequest(c, err.Error())
			return
		}
		if req.Format != "" && req.Format != "png" && req.Format != "jpeg" {
			apierr.BadRequest(c, "format must be png or jpeg")
			return
		}
		if req.Width > 4000 || req.Height > 8000 || req.Scale > 3 {
			apierr.BadRequest(c, "dimensions too large")
			return
		}

		img, err := rn.HTMLToImage(c.Request.Context(), req.HTML, render.Options{
			Width: req.Width, Height: req.Height, Scale: req.Scale,
			Format: req.Format, Quality: req.Quality,
		})
		if err != nil {
			apierr.Internal(c)
			return
		}

		ctype := "image/png"
		if req.Format == "jpeg" {
			ctype = "image/jpeg"
		}
		c.Data(http.StatusOK, ctype, img)
	}
}

// securityHeaders adds the basic protective headers to every response.
func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Frame-Options", "DENY")
		c.Next()
	}
}

// noStore keeps signed-in responses out of shared and browser caches.
func noStore() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "no-store")
		c.Next()
	}
}

// bodyLimit rejects oversized requests before anything is read.
func bodyLimit(max int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > max {
			apierr.PayloadTooLarge(c, "request body is too large")
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max)
		c.Next()
	}
}
