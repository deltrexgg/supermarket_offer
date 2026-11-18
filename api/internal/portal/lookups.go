package portal

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/deltrexgg/supermarket_offer/api/internal/apierr"
	"github.com/deltrexgg/supermarket_offer/api/internal/models"
)

// Areas and Categories are public: the portal forms and the public site filters use them.
func (h *Handler) Areas(c *gin.Context) {
	var rows []models.Area
	if err := h.DB.WithContext(c.Request.Context()).Order("name").Find(&rows).Error; err != nil {
		apierr.Internal(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}

func (h *Handler) Categories(c *gin.Context) {
	var rows []models.Category
	if err := h.DB.WithContext(c.Request.Context()).Order("name").Find(&rows).Error; err != nil {
		apierr.Internal(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}
