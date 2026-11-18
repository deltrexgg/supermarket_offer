// Package admin holds the HTTP handlers for the platform owner: dashboard counts,
// supermarket approval, area and category management, and offer removal.
// Every route here sits behind RequireRole("admin").
package admin

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/deltrexgg/supermarket_offer/api/internal/apierr"
	"github.com/deltrexgg/supermarket_offer/api/internal/cache"
	"github.com/deltrexgg/supermarket_offer/api/internal/models"
	"github.com/deltrexgg/supermarket_offer/api/internal/storage"
)

type Handler struct {
	DB    *gorm.DB
	Redis *redis.Client
	Store storage.Store
}

func pathID(c *gin.Context) (uint, bool) {
	n, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || n == 0 {
		return 0, false
	}
	return uint(n), true
}

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

func isUniqueViolation(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "23503"
}

// likePattern escapes LIKE wildcards in what the admin typed.
func likePattern(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(s) + "%"
}

// Stats: GET /admin/stats
func (h *Handler) Stats(c *gin.Context) {
	ctx := c.Request.Context()
	var out struct {
		SupermarketsPending  int64 `json:"supermarkets_pending"`
		SupermarketsApproved int64 `json:"supermarkets_approved"`
		SupermarketsBlocked  int64 `json:"supermarkets_blocked"`
		Branches             int64 `json:"branches"`
		OffersActive         int64 `json:"offers_active"`
		OffersUpcoming       int64 `json:"offers_upcoming"`
		OffersExpired        int64 `json:"offers_expired"`
		OffersEndingSoon     int64 `json:"offers_ending_soon"`
	}
	err := h.DB.WithContext(ctx).Raw(`
		SELECT
		  (SELECT count(*) FROM supermarkets WHERE status = 'pending')  AS supermarkets_pending,
		  (SELECT count(*) FROM supermarkets WHERE status = 'approved') AS supermarkets_approved,
		  (SELECT count(*) FROM supermarkets WHERE status = 'blocked')  AS supermarkets_blocked,
		  (SELECT count(*) FROM branches)                               AS branches,
		  (SELECT count(*) FROM offers_with_status WHERE status = 'active')   AS offers_active,
		  (SELECT count(*) FROM offers_with_status WHERE status = 'upcoming') AS offers_upcoming,
		  (SELECT count(*) FROM offers_with_status WHERE status = 'expired')  AS offers_expired,
		  (SELECT count(*) FROM offers_with_status WHERE status = 'active' AND days_left <= 2) AS offers_ending_soon
	`).Scan(&out).Error
	if err != nil {
		log.Printf("admin stats: %v", err)
		apierr.Internal(c)
		return
	}
	c.JSON(http.StatusOK, out)
}

// ---- supermarkets ----

type shopRow struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Phone       *string `json:"phone"`
	Status      string  `json:"status"`
	Email       string  `json:"email"`
	CreatedAt   string  `json:"created_at"`
	BranchCount int     `json:"branch_count"`
	OfferCount  int     `json:"offer_count"`
}

// ListSupermarkets: GET /admin/supermarkets?status=&q=&page=&page_size=
func (h *Handler) ListSupermarkets(c *gin.Context) {
	page, size := paging(c)
	f := apierr.Fields{}
	status := c.Query("status")
	if status != "" && status != "pending" && status != "approved" && status != "blocked" {
		f.Add("status", "must be pending, approved or blocked")
	}
	if !f.Empty() {
		apierr.Validation(c, f)
		return
	}

	q := h.DB.WithContext(c.Request.Context()).Table("supermarkets s")
	if status != "" {
		q = q.Where("s.status = ?", status)
	}
	if term := strings.TrimSpace(c.Query("q")); term != "" {
		q = q.Where("s.name ILIKE ?", likePattern(term))
	}
	q = q.Session(&gorm.Session{})

	var total int64
	if err := q.Count(&total).Error; err != nil {
		apierr.Internal(c)
		return
	}
	items := []shopRow{}
	err := q.Select(`s.id, s.name, s.slug, s.phone, s.status::text AS status,
			COALESCE((SELECT u.email::text FROM users u WHERE u.supermarket_id = s.id), '') AS email,
			to_char(s.created_at AT TIME ZONE 'Asia/Dubai', 'YYYY-MM-DD') AS created_at,
			(SELECT count(*) FROM branches b WHERE b.supermarket_id = s.id) AS branch_count,
			(SELECT count(*) FROM offers o WHERE o.supermarket_id = s.id) AS offer_count`).
		Order("CASE s.status WHEN 'pending' THEN 0 WHEN 'approved' THEN 1 ELSE 2 END, s.created_at DESC").
		Limit(size).Offset((page - 1) * size).Scan(&items).Error
	if err != nil {
		log.Printf("admin list supermarkets: %v", err)
		apierr.Internal(c)
		return
	}
	c.JSON(http.StatusOK, pageResult[shopRow]{Items: items, Total: total, Page: page, PageSize: size})
}

// SetSupermarketStatus: PATCH /admin/supermarkets/:id/status {"status":"approved|blocked|pending"}
func (h *Handler) SetSupermarketStatus(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		apierr.NotFound(c, "supermarket not found")
		return
	}
	var in struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		apierr.BadRequest(c, "invalid JSON body")
		return
	}
	if in.Status != "pending" && in.Status != "approved" && in.Status != "blocked" {
		apierr.Validation(c, map[string]string{"status": "must be pending, approved or blocked"})
		return
	}
	res := h.DB.WithContext(c.Request.Context()).Model(&models.Supermarket{}).
		Where("id = ?", id).Update("status", in.Status)
	if res.Error != nil {
		log.Printf("admin set status: %v", res.Error)
		apierr.Internal(c)
		return
	}
	if res.RowsAffected == 0 {
		apierr.NotFound(c, "supermarket not found")
		return
	}
	cache.BumpOffersVersion(c.Request.Context(), h.Redis)
	c.JSON(http.StatusOK, gin.H{"id": id, "status": in.Status})
}

// ---- areas and categories ----

type nameInput struct {
	Name string `json:"name"`
}

func readName(c *gin.Context) (string, bool) {
	var in nameInput
	if err := c.ShouldBindJSON(&in); err != nil {
		apierr.BadRequest(c, "invalid JSON body")
		return "", false
	}
	name := strings.TrimSpace(in.Name)
	if l := len([]rune(name)); l < 2 || l > 80 {
		apierr.Validation(c, map[string]string{"name": "must be 2 to 80 characters"})
		return "", false
	}
	return name, true
}

type lookupRow struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	InUse int    `json:"in_use"`
}

func (h *Handler) listLookup(c *gin.Context, table, usage string) {
	rows := []lookupRow{}
	err := h.DB.WithContext(c.Request.Context()).Raw(
		"SELECT t.id, t.name::text AS name, (SELECT count(*) FROM " + usage + " = t.id) AS in_use FROM " + table + " t ORDER BY t.name").
		Scan(&rows).Error
	if err != nil {
		apierr.Internal(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}

func (h *Handler) createLookup(c *gin.Context, table, label string) {
	name, ok := readName(c)
	if !ok {
		return
	}
	var id uint
	err := h.DB.WithContext(c.Request.Context()).Raw("INSERT INTO "+table+" (name) VALUES (?) RETURNING id", name).Scan(&id).Error
	if err != nil {
		if isUniqueViolation(err) {
			apierr.Conflict(c, label+" already exists")
			return
		}
		apierr.Internal(c)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "name": name})
}

func (h *Handler) renameLookup(c *gin.Context, table, label string) {
	id, ok := pathID(c)
	if !ok {
		apierr.NotFound(c, label+" not found")
		return
	}
	name, ok := readName(c)
	if !ok {
		return
	}
	res := h.DB.WithContext(c.Request.Context()).Exec("UPDATE "+table+" SET name = ? WHERE id = ?", name, id)
	if res.Error != nil {
		if isUniqueViolation(res.Error) {
			apierr.Conflict(c, label+" already exists")
			return
		}
		apierr.Internal(c)
		return
	}
	if res.RowsAffected == 0 {
		apierr.NotFound(c, label+" not found")
		return
	}
	cache.BumpOffersVersion(c.Request.Context(), h.Redis)
	c.JSON(http.StatusOK, gin.H{"id": id, "name": name})
}

func (h *Handler) deleteLookup(c *gin.Context, table, label string) {
	id, ok := pathID(c)
	if !ok {
		apierr.NotFound(c, label+" not found")
		return
	}
	res := h.DB.WithContext(c.Request.Context()).Exec("DELETE FROM "+table+" WHERE id = ?", id)
	if res.Error != nil {
		if isForeignKeyViolation(res.Error) {
			apierr.Conflict(c, label+" is still in use and cannot be deleted")
			return
		}
		apierr.Internal(c)
		return
	}
	if res.RowsAffected == 0 {
		apierr.NotFound(c, label+" not found")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ListAreas(c *gin.Context)  { h.listLookup(c, "areas", "branches WHERE area_id") }
func (h *Handler) CreateArea(c *gin.Context) { h.createLookup(c, "areas", "area") }
func (h *Handler) RenameArea(c *gin.Context) { h.renameLookup(c, "areas", "area") }
func (h *Handler) DeleteArea(c *gin.Context) { h.deleteLookup(c, "areas", "area") }

func (h *Handler) ListCategories(c *gin.Context) {
	h.listLookup(c, "categories", "offers WHERE category_id")
}
func (h *Handler) CreateCategory(c *gin.Context) { h.createLookup(c, "categories", "category") }
func (h *Handler) RenameCategory(c *gin.Context) { h.renameLookup(c, "categories", "category") }
func (h *Handler) DeleteCategory(c *gin.Context) { h.deleteLookup(c, "categories", "category") }
