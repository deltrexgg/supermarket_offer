package portal

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/deltrexgg/supermarket_offer/api/internal/ai"
	"github.com/deltrexgg/supermarket_offer/api/internal/apierr"
	"github.com/deltrexgg/supermarket_offer/api/internal/models"
)

type aiRequest struct {
	ItemName   string `json:"item_name"`
	Brand      string `json:"brand"`
	Unit       string `json:"unit"`
	CategoryID uint   `json:"category_id"`
}

// AIOfferText: POST /portal/ai/offer-text. It returns suggestions only; nothing is saved.
func (h *Handler) AIOfferText(c *gin.Context) {
	if h.AI == nil {
		apierr.ServiceUnavailable(c, "AI suggestions are not set up on this server")
		return
	}
	var req aiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid JSON body")
		return
	}
	req.ItemName = strings.TrimSpace(req.ItemName)
	req.Brand = strings.TrimSpace(req.Brand)
	req.Unit = strings.TrimSpace(req.Unit)
	f := apierr.Fields{}
	if l := len([]rune(req.ItemName)); l < 1 || l > 200 {
		f.Add("item_name", "is required (max 200 characters)")
	}
	if len([]rune(req.Brand)) > 100 {
		f.Add("brand", "must be at most 100 characters")
	}
	if len([]rune(req.Unit)) > 50 {
		f.Add("unit", "must be at most 50 characters")
	}
	if !f.Empty() {
		apierr.Validation(c, f)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 40*time.Second)
	defer cancel()

	in := ai.Input{ItemName: req.ItemName, Brand: req.Brand, Unit: req.Unit}
	if req.CategoryID != 0 {
		var cat models.Category
		if err := h.DB.WithContext(ctx).Where("id = ?", req.CategoryID).First(&cat).Error; err == nil {
			in.Category = cat.Name
		}
	}

	res, err := h.AI.OfferText(ctx, in)
	if err != nil {
		if errors.Is(err, ai.ErrNotConfigured) {
			apierr.ServiceUnavailable(c, "AI suggestions are not set up on this server")
			return
		}
		log.Printf("ai offer-text: %v", err)
		apierr.BadGateway(c, "the AI service could not write a suggestion right now, please try again")
		return
	}
	c.JSON(http.StatusOK, res)
}
