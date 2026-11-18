package portal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/deltrexgg/supermarket_offer/api/internal/apierr"
	"github.com/deltrexgg/supermarket_offer/api/internal/cache"
	"github.com/deltrexgg/supermarket_offer/api/internal/dubai"
	"github.com/deltrexgg/supermarket_offer/api/internal/models"
	"github.com/deltrexgg/supermarket_offer/api/internal/storage"
	"github.com/deltrexgg/supermarket_offer/api/internal/tenant"
)

type offerView struct {
	models.OfferWithStatus
	PhotoURL  *string `json:"photo_url"`
	PosterURL *string `json:"poster_url"`
	BranchIDs []uint  `json:"branch_ids"`
}

func (h *Handler) toView(o models.OfferWithStatus, branchIDs []uint) offerView {
	v := offerView{OfferWithStatus: o, BranchIDs: branchIDs}
	if v.BranchIDs == nil {
		v.BranchIDs = []uint{}
	}
	if o.PhotoPath != nil {
		u := h.Store.URL(*o.PhotoPath)
		v.PhotoURL = &u
	}
	if o.PosterPath != nil {
		u := h.Store.URL(*o.PosterPath)
		v.PosterURL = &u
	}
	return v
}

type offerInput struct {
	ItemName      string
	ItemNameAr    string
	Brand         string
	Description   string
	DescriptionAr string
	Unit          string
	CategoryID    uint
	Regular       float64
	Offer         float64
	ValidFrom     string
	ValidTo       string
	BranchIDs     []uint
}

// optional is a pointer for a create, nil when the text is empty.
func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// nullable is a value for an update map: SQL NULL when the text is empty.
func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func parsePrice(raw string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 || v > 1_000_000 {
		return 0, false
	}
	if math.Abs(v*100-math.Round(v*100)) > 1e-6 { // more than 2 decimals
		return 0, false
	}
	return math.Round(v*100) / 100, true
}

// readOfferForm reads and validates the multipart fields shared by create and update.
// Checks that need the database (category, branches) are done by checkOfferRefs.
func readOfferForm(c *gin.Context, requireFutureEnd bool) (offerInput, apierr.Fields) {
	f := apierr.Fields{}
	in := offerInput{
		ItemName:      strings.TrimSpace(c.PostForm("item_name")),
		ItemNameAr:    strings.TrimSpace(c.PostForm("item_name_ar")),
		Description:   strings.TrimSpace(c.PostForm("description")),
		DescriptionAr: strings.TrimSpace(c.PostForm("description_ar")),
		Brand:         strings.TrimSpace(c.PostForm("brand")),
		Unit:          strings.TrimSpace(c.PostForm("unit")),
		ValidFrom:     strings.TrimSpace(c.PostForm("valid_from")),
		ValidTo:       strings.TrimSpace(c.PostForm("valid_to")),
	}
	if l := len([]rune(in.ItemName)); l < 1 || l > 200 {
		f.Add("item_name", "is required (max 200 characters)")
	}
	if len([]rune(in.ItemNameAr)) > 200 {
		f.Add("item_name_ar", "must be at most 200 characters")
	}
	if len([]rune(in.Description)) > 1000 {
		f.Add("description", "must be at most 1000 characters")
	}
	if len([]rune(in.DescriptionAr)) > 1000 {
		f.Add("description_ar", "must be at most 1000 characters")
	}
	if len([]rune(in.Brand)) > 100 {
		f.Add("brand", "must be at most 100 characters")
	}
	if l := len([]rune(in.Unit)); l < 1 || l > 50 {
		f.Add("unit", "is required (max 50 characters, e.g. 1 L)")
	}
	if n, err := strconv.ParseUint(strings.TrimSpace(c.PostForm("category_id")), 10, 64); err != nil || n == 0 {
		f.Add("category_id", "is required")
	} else {
		in.CategoryID = uint(n)
	}

	reg, okReg := parsePrice(c.PostForm("regular_price"))
	off, okOff := parsePrice(c.PostForm("offer_price"))
	if !okReg {
		f.Add("regular_price", "must be a positive amount in AED with at most 2 decimals")
	}
	if !okOff {
		f.Add("offer_price", "must be a positive amount in AED with at most 2 decimals")
	}
	if okReg && okOff && off >= reg {
		f.Add("offer_price", "must be lower than the regular price")
	}
	in.Regular, in.Offer = reg, off

	fromOK, toOK := dubai.ValidDate(in.ValidFrom), dubai.ValidDate(in.ValidTo)
	if !fromOK {
		f.Add("valid_from", "must be a date like 2026-11-20")
	}
	if !toOK {
		f.Add("valid_to", "must be a date like 2026-11-27")
	}
	if fromOK && toOK {
		if in.ValidTo < in.ValidFrom {
			f.Add("valid_to", "must not be before the start date")
		} else if requireFutureEnd && in.ValidTo < dubai.Today() {
			f.Add("valid_to", "must not be in the past")
		}
	}

	raw := append(c.PostFormArray("branch_ids[]"), c.PostFormArray("branch_ids")...)
	seen := map[uint]bool{}
	for _, r := range raw {
		for _, part := range strings.Split(r, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			n, err := strconv.ParseUint(part, 10, 64)
			if err != nil || n == 0 {
				f.Add("branch_ids", "contains an invalid branch id")
				continue
			}
			if !seen[uint(n)] {
				seen[uint(n)] = true
				in.BranchIDs = append(in.BranchIDs, uint(n))
			}
		}
	}
	if len(in.BranchIDs) == 0 {
		f.Add("branch_ids", "choose at least one branch")
	}
	return in, f
}

// checkOfferRefs confirms the category exists and every branch belongs to this supermarket.
func (h *Handler) checkOfferRefs(ctx context.Context, smID uint, in offerInput, f apierr.Fields) error {
	var n int64
	if in.CategoryID != 0 {
		if err := h.DB.WithContext(ctx).Model(&models.Category{}).Where("id = ?", in.CategoryID).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			f.Add("category_id", "unknown category")
		}
	}
	if len(in.BranchIDs) > 0 {
		if err := h.DB.WithContext(ctx).Model(&models.Branch{}).
			Where("supermarket_id = ? AND id IN ?", smID, in.BranchIDs).Count(&n).Error; err != nil {
			return err
		}
		if int(n) != len(in.BranchIDs) {
			f.Add("branch_ids", "one or more branches are not yours")
		}
	}
	return nil
}

// readPhoto returns nil when no file was sent.
func readPhoto(c *gin.Context) (*storage.Image, bool) {
	fh, err := c.FormFile("photo")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return nil, true
		}
		apierr.BadRequest(c, "could not read the uploaded photo")
		return nil, false
	}
	return readImageHeader(c, fh)
}

func readImageHeader(c *gin.Context, fh *multipart.FileHeader) (*storage.Image, bool) {
	img, err := storage.ReadImage(fh)
	switch {
	case err == nil:
		return img, true
	case errors.Is(err, storage.ErrTooLarge):
		apierr.PayloadTooLarge(c, "the photo must be 5 MB or smaller")
	case errors.Is(err, storage.ErrUnsupported):
		apierr.UnsupportedMedia(c, "the photo must be a JPEG, PNG, WebP, GIF or AVIF image")
	default:
		apierr.BadRequest(c, "could not read the uploaded photo")
	}
	return nil, false
}

func photoKey(smID uint, ext string) string {
	return fmt.Sprintf("offers/%d/%s%s", smID, uuid.NewString(), ext)
}

func (h *Handler) removeFile(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := h.Store.Delete(ctx, key); err != nil {
		log.Printf("storage cleanup: %v", err)
	}
}

func (h *Handler) loadOffer(ctx context.Context, smID uint, id uint) (*models.OfferWithStatus, error) {
	var o models.OfferWithStatus
	err := h.DB.WithContext(ctx).Scopes(tenant.Scope(&smID)).Where("id = ?", id).First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (h *Handler) branchIDsFor(ctx context.Context, offerIDs []uint) (map[uint][]uint, error) {
	out := map[uint][]uint{}
	if len(offerIDs) == 0 {
		return out, nil
	}
	var links []models.OfferBranch
	if err := h.DB.WithContext(ctx).Where("offer_id IN ?", offerIDs).Find(&links).Error; err != nil {
		return nil, err
	}
	for _, l := range links {
		out[l.OfferID] = append(out[l.OfferID], l.BranchID)
	}
	for k := range out {
		sort.Slice(out[k], func(i, j int) bool { return out[k][i] < out[k][j] })
	}
	return out, nil
}

func (h *Handler) CreateOffer(c *gin.Context) {
	ctx := c.Request.Context()
	smID := *supermarketID(c)

	in, f := readOfferForm(c, true)
	if err := h.checkOfferRefs(ctx, smID, in, f); err != nil {
		apierr.Internal(c)
		return
	}
	if !f.Empty() {
		apierr.Validation(c, f)
		return
	}
	img, ok := readPhoto(c)
	if !ok {
		return
	}

	offer := models.Offer{
		SupermarketID: smID, CategoryID: in.CategoryID,
		ItemName: in.ItemName, Unit: in.Unit,
		RegularPrice: in.Regular, OfferPrice: in.Offer,
		ValidFrom: in.ValidFrom, ValidTo: in.ValidTo,
	}
	if in.Brand != "" {
		b := in.Brand
		offer.Brand = &b
	}
	offer.ItemNameAr = optional(in.ItemNameAr)
	offer.Description = optional(in.Description)
	offer.DescriptionAr = optional(in.DescriptionAr)
	var key string
	if img != nil {
		key = photoKey(smID, img.Ext)
		if err := h.Store.Put(ctx, key, img.ContentType, img.Data); err != nil {
			log.Printf("create offer: %v", err)
			apierr.Internal(c)
			return
		}
		offer.PhotoPath = &key
	}

	err := h.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&offer).Error; err != nil {
			return err
		}
		links := make([]models.OfferBranch, 0, len(in.BranchIDs))
		for _, b := range in.BranchIDs {
			links = append(links, models.OfferBranch{OfferID: offer.ID, BranchID: b, SupermarketID: smID})
		}
		return tx.Create(&links).Error
	})
	if err != nil {
		if key != "" {
			h.removeFile(key)
		}
		log.Printf("create offer: %v", err)
		apierr.Internal(c)
		return
	}

	cache.BumpOffersVersion(ctx, h.Redis)
	o, err := h.loadOffer(ctx, smID, offer.ID)
	if err != nil {
		apierr.Internal(c)
		return
	}
	c.JSON(http.StatusCreated, h.toView(*o, in.BranchIDs))
}

func (h *Handler) ListOffers(c *gin.Context) {
	ctx := c.Request.Context()
	smID := supermarketID(c)
	page, size := paging(c)

	q := h.DB.WithContext(ctx).Model(&models.OfferWithStatus{}).Scopes(tenant.Scope(smID))
	switch s := c.Query("status"); s {
	case "":
	case "upcoming", "active", "expired":
		q = q.Where("status = ?", s)
	default:
		apierr.BadRequest(c, "status must be upcoming, active or expired")
		return
	}
	if term := strings.TrimSpace(c.Query("q")); term != "" {
		like := "%" + strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(term) + "%"
		q = q.Where("(item_name ILIKE ? OR brand ILIKE ?)", like, like)
	}

	q = q.Session(&gorm.Session{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		apierr.Internal(c)
		return
	}
	var rows []models.OfferWithStatus
	if err := q.Order("id DESC").Limit(size).Offset((page - 1) * size).Find(&rows).Error; err != nil {
		apierr.Internal(c)
		return
	}
	ids := make([]uint, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	byOffer, err := h.branchIDsFor(ctx, ids)
	if err != nil {
		apierr.Internal(c)
		return
	}
	items := make([]offerView, 0, len(rows))
	for _, r := range rows {
		items = append(items, h.toView(r, byOffer[r.ID]))
	}
	c.JSON(http.StatusOK, pageResult[offerView]{Items: items, Total: total, Page: page, PageSize: size})
}

func (h *Handler) GetOffer(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		apierr.NotFound(c, "offer not found")
		return
	}
	ctx := c.Request.Context()
	o, err := h.loadOffer(ctx, *supermarketID(c), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "offer not found")
			return
		}
		apierr.Internal(c)
		return
	}
	byOffer, err := h.branchIDsFor(ctx, []uint{o.ID})
	if err != nil {
		apierr.Internal(c)
		return
	}
	c.JSON(http.StatusOK, h.toView(*o, byOffer[o.ID]))
}

// UpdateOffer replaces the offer's fields and branch list. A new photo replaces
// the old one; sending no photo keeps the current one.
func (h *Handler) UpdateOffer(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		apierr.NotFound(c, "offer not found")
		return
	}
	ctx := c.Request.Context()
	smID := *supermarketID(c)

	current, err := h.loadOffer(ctx, smID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "offer not found")
			return
		}
		apierr.Internal(c)
		return
	}

	in, f := readOfferForm(c, false)
	if err := h.checkOfferRefs(ctx, smID, in, f); err != nil {
		apierr.Internal(c)
		return
	}
	if !f.Empty() {
		apierr.Validation(c, f)
		return
	}
	img, ok := readPhoto(c)
	if !ok {
		return
	}

	var brand any
	if in.Brand != "" {
		brand = in.Brand
	}
	updates := map[string]any{
		"category_id": in.CategoryID, "item_name": in.ItemName, "brand": brand, "unit": in.Unit,
		"regular_price": in.Regular, "offer_price": in.Offer,
		"valid_from": in.ValidFrom, "valid_to": in.ValidTo,
		"item_name_ar": nullable(in.ItemNameAr), "description": nullable(in.Description),
		"description_ar": nullable(in.DescriptionAr),
	}
	// Any edit makes the saved poster out of date, so it is cleared.
	updates["poster_path"] = nil
	var newKey string
	if img != nil {
		newKey = photoKey(smID, img.Ext)
		if err := h.Store.Put(ctx, newKey, img.ContentType, img.Data); err != nil {
			log.Printf("update offer: %v", err)
			apierr.Internal(c)
			return
		}
		updates["photo_path"] = newKey
	}

	err = h.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Offer{}).Scopes(tenant.Scope(&smID)).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Where("offer_id = ?", id).Delete(&models.OfferBranch{}).Error; err != nil {
			return err
		}
		links := make([]models.OfferBranch, 0, len(in.BranchIDs))
		for _, b := range in.BranchIDs {
			links = append(links, models.OfferBranch{OfferID: id, BranchID: b, SupermarketID: smID})
		}
		return tx.Create(&links).Error
	})
	if err != nil {
		if newKey != "" {
			h.removeFile(newKey)
		}
		log.Printf("update offer: %v", err)
		apierr.Internal(c)
		return
	}
	if newKey != "" && current.PhotoPath != nil {
		h.removeFile(*current.PhotoPath)
	}
	if current.PosterPath != nil {
		h.removeFile(*current.PosterPath)
	}

	cache.BumpOffersVersion(ctx, h.Redis)
	o, err := h.loadOffer(ctx, smID, id)
	if err != nil {
		apierr.Internal(c)
		return
	}
	c.JSON(http.StatusOK, h.toView(*o, in.BranchIDs))
}

func (h *Handler) DeleteOffer(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		apierr.NotFound(c, "offer not found")
		return
	}
	ctx := c.Request.Context()
	smID := *supermarketID(c)

	o, err := h.loadOffer(ctx, smID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "offer not found")
			return
		}
		apierr.Internal(c)
		return
	}
	if err := h.DB.WithContext(ctx).Scopes(tenant.Scope(&smID)).Where("id = ?", id).Delete(&models.Offer{}).Error; err != nil {
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
