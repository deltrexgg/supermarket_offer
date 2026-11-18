package portal

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/deltrexgg/supermarket_offer/api/internal/apierr"
	"github.com/deltrexgg/supermarket_offer/api/internal/models"
	"github.com/deltrexgg/supermarket_offer/api/internal/poster"
	"github.com/deltrexgg/supermarket_offer/api/internal/render"
	"github.com/deltrexgg/supermarket_offer/api/internal/tenant"
)

// Chrome is heavy: allow only a couple of renders at once, the rest wait.
var renderSlots = make(chan struct{}, 2)

type posterRequest struct {
	Template string `json:"template"`
	Lang     string `json:"lang"`
}

func date10(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

// posterHTML loads everything the poster shows and builds its HTML page.
// It returns false after writing an error response.
func (h *Handler) posterHTML(c *gin.Context, o *models.OfferWithStatus, tpl, lang string, print bool) (string, bool) {
	ctx := c.Request.Context()
	smID := *supermarketID(c)

	var sm models.Supermarket
	if err := h.DB.WithContext(ctx).First(&sm, smID).Error; err != nil {
		apierr.Internal(c)
		return "", false
	}

	var rows []struct{ Name, Area string }
	if err := h.DB.WithContext(ctx).Table("offer_branches ob").
		Select("b.name AS name, a.name AS area").
		Joins("JOIN branches b ON b.id = ob.branch_id").
		Joins("JOIN areas a ON a.id = b.area_id").
		Where("ob.offer_id = ? AND ob.supermarket_id = ?", o.ID, smID).
		Order("b.name").Scan(&rows).Error; err != nil {
		apierr.Internal(c)
		return "", false
	}
	shop := poster.Shop{Name: sm.Name}
	seenArea := map[string]bool{}
	for _, r := range rows {
		shop.Branches = append(shop.Branches, r.Name)
		if !seenArea[r.Area] {
			seenArea[r.Area] = true
			shop.Areas = append(shop.Areas, r.Area)
		}
	}

	po := poster.Offer{
		ItemName: o.ItemName, Unit: o.Unit,
		RegularPrice: o.RegularPrice, OfferPrice: o.OfferPrice, DiscountPct: int(o.DiscountPct),
		ValidFrom: date10(o.ValidFrom), ValidTo: date10(o.ValidTo),
		Status: o.Status, DaysLeft: o.DaysLeft,
	}
	if o.ItemNameAr != nil {
		po.ItemNameAr = *o.ItemNameAr
	}
	if o.Brand != nil {
		po.Brand = *o.Brand
	}
	if o.PhotoPath != nil {
		// A missing photo must not block the poster: it falls back to a placeholder.
		if data, ct, err := h.Store.Get(ctx, *o.PhotoPath); err != nil {
			log.Printf("poster: photo unavailable: %v", err)
		} else {
			po.PhotoBytes, po.PhotoType = data, ct
		}
	}

	html, err := poster.HTML(tpl, lang, po, shop, print)
	if err != nil {
		apierr.Internal(c)
		return "", false
	}
	return html, true
}

func (h *Handler) offerForPoster(c *gin.Context) (*models.OfferWithStatus, bool) {
	id, ok := pathID(c)
	if !ok {
		apierr.NotFound(c, "offer not found")
		return nil, false
	}
	o, err := h.loadOffer(c.Request.Context(), *supermarketID(c), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "offer not found")
		} else {
			apierr.Internal(c)
		}
		return nil, false
	}
	return o, true
}

// CreatePoster renders the offer to an A4 PNG, stores it and saves its path.
// Every render gets a new file name, so browsers and CDNs never show an old poster.
func (h *Handler) CreatePoster(c *gin.Context) {
	var req posterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "body must be JSON")
		return
	}
	if req.Template == "" {
		req.Template = poster.HotDeal
	}
	if req.Lang == "" {
		req.Lang = poster.LangEN
	}
	if !poster.ValidChoice(req.Template, req.Lang) {
		f := apierr.Fields{}
		if req.Template != poster.HotDeal && req.Template != poster.WeekendOffer {
			f.Add("template", "must be hot_deal or weekend_offer")
		}
		if req.Lang != poster.LangEN && req.Lang != poster.LangAR {
			f.Add("lang", "must be en or ar")
		}
		apierr.Validation(c, f)
		return
	}

	o, ok := h.offerForPoster(c)
	if !ok {
		return
	}
	html, ok := h.posterHTML(c, o, req.Template, req.Lang, false)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	select {
	case renderSlots <- struct{}{}:
		defer func() { <-renderSlots }()
	case <-ctx.Done():
		apierr.Internal(c)
		return
	}
	img, err := h.Renderer.HTMLToImage(ctx, html, render.Options{
		Width: poster.Width, Height: poster.Height, Scale: 1, Format: "png",
	})
	if err != nil {
		log.Printf("poster render: %v", err)
		apierr.Internal(c)
		return
	}

	smID := *supermarketID(c)
	key := fmt.Sprintf("posters/%d-%d.png", o.ID, time.Now().Unix())
	if err := h.Store.Put(ctx, key, "image/png", img); err != nil {
		log.Printf("poster store: %v", err)
		apierr.Internal(c)
		return
	}
	res := h.DB.WithContext(ctx).Model(&models.Offer{}).Scopes(tenant.Scope(&smID)).
		Where("id = ?", o.ID).Update("poster_path", key)
	if res.Error != nil || res.RowsAffected == 0 {
		h.removeFile(key)
		apierr.Internal(c)
		return
	}
	if o.PosterPath != nil {
		h.removeFile(*o.PosterPath)
	}
	c.JSON(http.StatusOK, gin.H{"poster_url": h.Store.URL(key)})
}

// DownloadPoster sends the saved PNG as a file download.
func (h *Handler) DownloadPoster(c *gin.Context) {
	o, ok := h.offerForPoster(c)
	if !ok {
		return
	}
	if o.PosterPath == nil {
		apierr.NotFound(c, "no poster yet: create one first")
		return
	}
	data, _, err := h.Store.Get(c.Request.Context(), *o.PosterPath)
	if err != nil {
		log.Printf("poster download: %v", err)
		apierr.Internal(c)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="poster-%d.png"`, o.ID))
	c.Data(http.StatusOK, "image/png", data)
}

// PrintPoster returns the poster as an A4 page that opens the print dialog.
func (h *Handler) PrintPoster(c *gin.Context) {
	tpl, lang := c.DefaultQuery("template", poster.HotDeal), c.DefaultQuery("lang", poster.LangEN)
	if !poster.ValidChoice(tpl, lang) {
		apierr.BadRequest(c, "template must be hot_deal or weekend_offer and lang must be en or ar")
		return
	}
	o, ok := h.offerForPoster(c)
	if !ok {
		return
	}
	html, ok := h.posterHTML(c, o, tpl, lang, true)
	if !ok {
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
