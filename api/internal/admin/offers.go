package admin

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/deltrexgg/supermarket_offer/api/internal/apierr"
	"github.com/deltrexgg/supermarket_offer/api/internal/cache"
	"github.com/deltrexgg/supermarket_offer/api/internal/models"
)

type offerRow struct {
	ID              uint    `json:"id"`
	ItemName        string  `json:"item_name"`
	Brand           *string `json:"brand"`
	Unit            string  `json:"unit"`
	RegularPrice    float64 `json:"regular_price"`
	OfferPrice      float64 `json:"offer_price"`
	DiscountPct     int     `json:"discount_pct"`
	ValidFrom       string  `json:"valid_from"`
	ValidTo         string  `json:"valid_to"`
	Status          string  `json:"status"`
	DaysLeft        int     `json:"days_left"`
	SupermarketID   uint    `json:"supermarket_id"`
	SupermarketName string  `json:"supermarket_name"`
}

// ListOffers: GET /admin/offers?status=&supermarket=&q=&page=&page_size=
// Shows offers of every supermarket, whatever their dates or the shop's status.
func (h *Handler) ListOffers(c *gin.Context) {
	page, size := paging(c)
	f := apierr.Fields{}
	status := c.Query("status")
	if status != "" && status != "active" && status != "upcoming" && status != "expired" {
		f.Add("status", "must be active, upcoming or expired")
	}
	var shop uint64
	if v := c.Query("supermarket"); v != "" {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil || n == 0 {
			f.Add("supermarket", "must be a supermarket id")
		}
		shop = n
	}
	if !f.Empty() {
		apierr.Validation(c, f)
		return
	}

	q := h.DB.WithContext(c.Request.Context()).Table("offers_with_status o").
		Joins("JOIN supermarkets s ON s.id = o.supermarket_id")
	if status != "" {
		q = q.Where("o.status = ?", status)
	}
	if shop > 0 {
		q = q.Where("o.supermarket_id = ?", shop)
	}
	if term := strings.TrimSpace(c.Query("q")); term != "" {
		q = q.Where("(o.item_name ILIKE ? OR o.brand ILIKE ?)", likePattern(term), likePattern(term))
	}
	q = q.Session(&gorm.Session{})

	var total int64
	if err := q.Count(&total).Error; err != nil {
		apierr.Internal(c)
		return
	}
	items := []offerRow{}
	err := q.Select(`o.id, o.item_name, o.brand, o.unit, o.regular_price, o.offer_price, o.discount_pct,
			to_char(o.valid_from, 'YYYY-MM-DD') AS valid_from, to_char(o.valid_to, 'YYYY-MM-DD') AS valid_to,
			o.status, o.days_left, o.supermarket_id, s.name AS supermarket_name`).
		Order("o.id DESC").Limit(size).Offset((page - 1) * size).Scan(&items).Error
	if err != nil {
		log.Printf("admin list offers: %v", err)
		apierr.Internal(c)
		return
	}
	c.JSON(http.StatusOK, pageResult[offerRow]{Items: items, Total: total, Page: page, PageSize: size})
}

// DeleteOffer: DELETE /admin/offers/:id removes an offer and its stored files.
func (h *Handler) DeleteOffer(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		apierr.NotFound(c, "offer not found")
		return
	}
	ctx := c.Request.Context()
	var o models.Offer
	if err := h.DB.WithContext(ctx).Where("id = ?", id).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "offer not found")
			return
		}
		apierr.Internal(c)
		return
	}
	if err := h.DB.WithContext(ctx).Where("id = ?", id).Delete(&models.Offer{}).Error; err != nil {
		apierr.Internal(c)
		return
	}
	for _, p := range []*string{o.PhotoPath, o.PosterPath} {
		if p != nil {
			h.removeFile(*p)
		}
	}
	cache.BumpOffersVersion(ctx, h.Redis)
	c.Status(http.StatusNoContent)
}

// removeFile is best effort: a leftover file is harmless, a failed request is not.
func (h *Handler) removeFile(key string) {
	if h.Store == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := h.Store.Delete(ctx, key); err != nil {
		log.Printf("admin: delete file %s: %v", key, err)
	}
}
