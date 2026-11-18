package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/deltrexgg/supermarket_offer/api/internal/apierr"
	"github.com/deltrexgg/supermarket_offer/api/internal/auth"
	"github.com/deltrexgg/supermarket_offer/api/internal/render"
)

type Deps struct {
	DB          *gorm.DB
	Redis       *redis.Client
	Renderer    *render.Renderer
	Auth        *auth.Verifier
	CORSOrigins string
}

func NewRouter(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
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

	// Signed-in routes. Phase 2+ adds the supermarket and admin groups here.
	authed := v1.Group("", d.Auth.Authenticate())
	authed.GET("/auth/me", func(c *gin.Context) { c.JSON(http.StatusOK, auth.CurrentUser(c)) })

	// Rendering runs headless Chrome, so only signed-in users may call it.
	authed.POST("/render", renderHandler(d.Renderer))

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
