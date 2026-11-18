// Package portal holds the HTTP handlers for supermarket owners: registration,
// branches and offers. Every query is scoped to the signed-in user's supermarket.
package portal

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/deltrexgg/supermarket_offer/api/internal/auth"
	"github.com/deltrexgg/supermarket_offer/api/internal/render"
	"github.com/deltrexgg/supermarket_offer/api/internal/storage"
	"github.com/deltrexgg/supermarket_offer/api/internal/supabase"
)

type Handler struct {
	DB       *gorm.DB
	Redis    *redis.Client
	Store    storage.Store
	Admin    *supabase.Admin
	Renderer *render.Renderer
}

// supermarketID is the tenant of the request. RequireRole("supermarket") runs
// first, and the database guarantees such users always have a supermarket.
func supermarketID(c *gin.Context) *uint {
	if u := auth.CurrentUser(c); u != nil {
		return u.SupermarketID
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "23503"
}

func pathID(c *gin.Context) (uint, bool) {
	n, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || n == 0 {
		return 0, false
	}
	return uint(n), true
}

// paging reads page and page_size (default 20, max 50).
func paging(c *gin.Context) (page, size int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 50 {
		size = 50
	}
	return page, size
}

type pageResult[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}
