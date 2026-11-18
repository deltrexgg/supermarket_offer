// Package public serves the website's read-only API: active offers from approved
// supermarkets, as of today's date in Dubai. Responses are cached in Redis until
// Dubai midnight, and any portal change bumps a version number that retires them.
package public

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/deltrexgg/supermarket_offer/api/internal/apierr"
	"github.com/deltrexgg/supermarket_offer/api/internal/cache"
	"github.com/deltrexgg/supermarket_offer/api/internal/dubai"
	"github.com/deltrexgg/supermarket_offer/api/internal/storage"
)

type Handler struct {
	DB    *gorm.DB
	Redis *redis.Client
	Store storage.Store
}

var sortOrders = map[string]string{
	"lowest_price":     "o.offer_price ASC, o.id DESC",
	"highest_discount": "o.discount_pct DESC, o.id DESC",
	"ending_soon":      "o.valid_to ASC, o.id DESC",
}

const defaultSort = "highest_discount"

type Card struct {
	ID              uint    `json:"id"`
	ItemName        string  `json:"item_name"`
	ItemNameAr      *string `json:"item_name_ar"`
	Brand           *string `json:"brand"`
	Unit            string  `json:"unit"`
	RegularPrice    float64 `json:"regular_price"`
	OfferPrice      float64 `json:"offer_price"`
	DiscountPct     int     `json:"discount_pct"`
	ValidTo         string  `json:"valid_to"`
	DaysLeft        int     `json:"days_left"`
	PhotoURL        *string `json:"photo_url"`
	SupermarketID   uint    `json:"supermarket_id"`
	SupermarketName string  `json:"supermarket_name"`
	LogoURL         *string `json:"logo_url"`
	BranchCount     int     `json:"branch_count"`
}

type CardRow struct {
	ID              uint
	ItemName        string
	ItemNameAr      *string
	Brand           *string
	Unit            string
	RegularPrice    float64
	OfferPrice      float64
	DiscountPct     int
	ValidTo         string
	DaysLeft        int
	PhotoPath       *string
	SupermarketID   uint
	SupermarketName string
	LogoPath        *string
	BranchCount     int
}

func (h *Handler) toCard(r CardRow) Card {
	c := Card{
		ID: r.ID, ItemName: r.ItemName, ItemNameAr: r.ItemNameAr, Brand: r.Brand, Unit: r.Unit,
		RegularPrice: r.RegularPrice, OfferPrice: r.OfferPrice, DiscountPct: r.DiscountPct,
		ValidTo: r.ValidTo, DaysLeft: r.DaysLeft, SupermarketID: r.SupermarketID,
		SupermarketName: r.SupermarketName, BranchCount: r.BranchCount,
	}
	c.PhotoURL = h.fileURL(r.PhotoPath)
	c.LogoURL = h.fileURL(r.LogoPath)
	return c
}

func (h *Handler) fileURL(p *string) *string {
	if p == nil || *p == "" {
		return nil
	}
	u := h.Store.URL(*p)
	return &u
}

func positiveInt(s string) (uint, bool) {
	n, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
	return uint(n), err == nil && n > 0
}

// cacheKey is offers:v{version}:{dubai date}:{name}:{hash of the normalised query}.
// The date changes at Dubai midnight, the version after every portal change.
func (h *Handler) cacheKey(ctx context.Context, name string, q url.Values) string {
	return "offers:v" + cache.OffersVersion(ctx, h.Redis) + ":" + dubai.Today() + ":" + name + ":" + queryHash(q)
}

// queryHash is the same for the same query whatever the order of its parameters.
func queryHash(q url.Values) string {
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k + "=" + strings.Join(q[k], ",") + "&")
	}
	sum := sha1.Sum([]byte(b.String()))
	return hex.EncodeToString(sum[:8])
}

func (h *Handler) respond(c *gin.Context, key string, build func() ([]byte, error)) {
	body, hit, err := cache.Remember(c.Request.Context(), h.Redis, key, dubai.UntilMidnight(), build)
	if err != nil {
		if errors.Is(err, errNotFound) {
			apierr.NotFound(c, "offer not found")
			return
		}
		apierr.Internal(c)
		return
	}
	if hit {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

var errNotFound = errors.New("not found")

func marshal(v any) ([]byte, error) { return json.Marshal(v) }
