package admin

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/deltrexgg/supermarket_offer/api/internal/apierr"
	"github.com/deltrexgg/supermarket_offer/api/internal/cache"
	"github.com/deltrexgg/supermarket_offer/api/internal/logo"
	"github.com/deltrexgg/supermarket_offer/api/internal/models"
)

// SetSupermarketLogo: POST /admin/supermarkets/:id/logo, multipart with "logo".
func (h *Handler) SetSupermarketLogo(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		apierr.NotFound(c, "supermarket not found")
		return
	}
	img, ok := logo.Read(c)
	if !ok {
		return
	}
	if img == nil {
		apierr.BadRequest(c, "attach the logo as a file field named logo")
		return
	}

	ctx := c.Request.Context()
	var sm models.Supermarket
	if err := h.DB.WithContext(ctx).First(&sm, id).Error; err != nil {
		apierr.NotFound(c, "supermarket not found")
		return
	}

	key := logo.Key(id, img.Ext)
	if err := h.Store.Put(ctx, key, img.ContentType, img.Data); err != nil {
		log.Printf("admin logo upload: %v", err)
		apierr.BadGateway(c, "could not store the logo")
		return
	}
	if err := h.DB.WithContext(ctx).Model(&models.Supermarket{}).
		Where("id = ?", id).Update("logo_path", key).Error; err != nil {
		logo.RemoveFile(h.Store, key)
		log.Printf("admin logo save: %v", err)
		apierr.Internal(c)
		return
	}
	if sm.LogoPath != nil && *sm.LogoPath != key && logo.Owned(id, *sm.LogoPath) {
		logo.RemoveFile(h.Store, *sm.LogoPath)
	}
	cache.BumpOffersVersion(ctx, h.Redis)
	c.JSON(http.StatusOK, gin.H{"id": id, "logo_url": h.Store.URL(key)})
}

// RemoveSupermarketLogo: DELETE /admin/supermarkets/:id/logo.
func (h *Handler) RemoveSupermarketLogo(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		apierr.NotFound(c, "supermarket not found")
		return
	}
	ctx := c.Request.Context()
	var sm models.Supermarket
	if err := h.DB.WithContext(ctx).First(&sm, id).Error; err != nil {
		apierr.NotFound(c, "supermarket not found")
		return
	}
	if err := h.DB.WithContext(ctx).Model(&models.Supermarket{}).
		Where("id = ?", id).Update("logo_path", nil).Error; err != nil {
		log.Printf("admin logo remove: %v", err)
		apierr.Internal(c)
		return
	}
	if sm.LogoPath != nil && logo.Owned(id, *sm.LogoPath) {
		logo.RemoveFile(h.Store, *sm.LogoPath)
	}
	cache.BumpOffersVersion(ctx, h.Redis)
	c.JSON(http.StatusOK, gin.H{"id": id, "logo_url": nil})
}
