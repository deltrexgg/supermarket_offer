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

	"github.com/deltrexgg/supermarket_offer/api/internal/render"
)

type Deps struct {
	DB          *gorm.DB
	Redis       *redis.Client
	Renderer    *render.Renderer
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

	v1 := r.Group("/api/v1")
	v1.POST("/render", renderHandler(d.Renderer))

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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Format != "" && req.Format != "png" && req.Format != "jpeg" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "format must be png or jpeg"})
			return
		}
		if req.Width > 4000 || req.Height > 8000 || req.Scale > 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "dimensions too large"})
			return
		}

		img, err := rn.HTMLToImage(c.Request.Context(), req.HTML, render.Options{
			Width: req.Width, Height: req.Height, Scale: req.Scale,
			Format: req.Format, Quality: req.Quality,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctype := "image/png"
		if req.Format == "jpeg" {
			ctype = "image/jpeg"
		}
		c.Data(http.StatusOK, ctype, img)
	}
}
