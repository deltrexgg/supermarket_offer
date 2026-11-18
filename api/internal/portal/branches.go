package portal

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/deltrexgg/supermarket_offer/api/internal/apierr"
	"github.com/deltrexgg/supermarket_offer/api/internal/cache"
	"github.com/deltrexgg/supermarket_offer/api/internal/models"
	"github.com/deltrexgg/supermarket_offer/api/internal/tenant"
)

type branchView struct {
	models.Branch
	AreaName string `json:"area_name"`
}

type branchRequest struct {
	Name    string `json:"name"`
	AreaID  uint   `json:"area_id"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

func (h *Handler) branchQuery(c *gin.Context) *gorm.DB {
	return h.DB.WithContext(c.Request.Context()).
		Table("branches").
		Select("branches.*, areas.name AS area_name").
		Joins("JOIN areas ON areas.id = branches.area_id").
		Scopes(tenant.Scope(supermarketID(c)))
}

func (h *Handler) validateBranch(c *gin.Context, req *branchRequest) bool {
	req.Name = strings.TrimSpace(req.Name)
	req.Address = strings.TrimSpace(req.Address)
	req.Phone = strings.TrimSpace(req.Phone)
	f := apierr.Fields{}
	if l := len([]rune(req.Name)); l < 1 || l > 150 {
		f.Add("name", "is required (max 150 characters)")
	}
	if req.Address == "" {
		f.Add("address", "is required")
	}
	if len(req.Phone) > 30 {
		f.Add("phone", "must be at most 30 characters")
	}
	if req.AreaID == 0 {
		f.Add("area_id", "is required")
	} else {
		var n int64
		if err := h.DB.WithContext(c.Request.Context()).Model(&models.Area{}).Where("id = ?", req.AreaID).Count(&n).Error; err != nil {
			apierr.Internal(c)
			return false
		}
		if n == 0 {
			f.Add("area_id", "unknown area")
		}
	}
	if !f.Empty() {
		apierr.Validation(c, f)
		return false
	}
	return true
}

func (h *Handler) ListBranches(c *gin.Context) {
	var rows []branchView
	if err := h.branchQuery(c).Order("branches.name").Scan(&rows).Error; err != nil {
		apierr.Internal(c)
		return
	}
	if rows == nil {
		rows = []branchView{}
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}

func (h *Handler) loadBranch(c *gin.Context, id uint) (*branchView, bool) {
	var rows []branchView
	if err := h.branchQuery(c).Where("branches.id = ?", id).Scan(&rows).Error; err != nil {
		apierr.Internal(c)
		return nil, false
	}
	if len(rows) == 0 {
		apierr.NotFound(c, "branch not found")
		return nil, false
	}
	return &rows[0], true
}

func (h *Handler) GetBranch(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		apierr.NotFound(c, "branch not found")
		return
	}
	if b, ok := h.loadBranch(c, id); ok {
		c.JSON(http.StatusOK, b)
	}
}

func (h *Handler) CreateBranch(c *gin.Context) {
	var req branchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "body must be JSON")
		return
	}
	if !h.validateBranch(c, &req) {
		return
	}
	b := models.Branch{SupermarketID: *supermarketID(c), AreaID: req.AreaID, Name: req.Name, Address: req.Address}
	if req.Phone != "" {
		p := req.Phone
		b.Phone = &p
	}
	if err := h.DB.WithContext(c.Request.Context()).Create(&b).Error; err != nil {
		if isUniqueViolation(err) {
			apierr.Conflict(c, "you already have a branch with this name")
			return
		}
		apierr.Internal(c)
		return
	}
	cache.BumpOffersVersion(c.Request.Context(), h.Redis)
	if v, ok := h.loadBranch(c, b.ID); ok {
		c.JSON(http.StatusCreated, v)
	}
}

func (h *Handler) UpdateBranch(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		apierr.NotFound(c, "branch not found")
		return
	}
	var req branchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "body must be JSON")
		return
	}
	if !h.validateBranch(c, &req) {
		return
	}
	var phone any
	if req.Phone != "" {
		phone = req.Phone
	}
	res := h.DB.WithContext(c.Request.Context()).Model(&models.Branch{}).
		Scopes(tenant.Scope(supermarketID(c))).
		Where("id = ?", id).
		Updates(map[string]any{"name": req.Name, "area_id": req.AreaID, "address": req.Address, "phone": phone})
	if res.Error != nil {
		if isUniqueViolation(res.Error) {
			apierr.Conflict(c, "you already have a branch with this name")
			return
		}
		apierr.Internal(c)
		return
	}
	if res.RowsAffected == 0 {
		apierr.NotFound(c, "branch not found")
		return
	}
	cache.BumpOffersVersion(c.Request.Context(), h.Redis)
	if v, ok := h.loadBranch(c, id); ok {
		c.JSON(http.StatusOK, v)
	}
}

// DeleteBranch refuses while offers still use the branch, so no offer is left
// without a place where it can be bought.
func (h *Handler) DeleteBranch(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		apierr.NotFound(c, "branch not found")
		return
	}
	ctx := c.Request.Context()
	if _, ok := h.loadBranch(c, id); !ok {
		return
	}
	var used int64
	if err := h.DB.WithContext(ctx).Model(&models.OfferBranch{}).Where("branch_id = ?", id).Count(&used).Error; err != nil {
		apierr.Internal(c)
		return
	}
	if used > 0 {
		apierr.Conflict(c, "this branch is used by offers: remove it from them first")
		return
	}
	res := h.DB.WithContext(ctx).Scopes(tenant.Scope(supermarketID(c))).Where("id = ?", id).Delete(&models.Branch{})
	if res.Error != nil {
		if isForeignKeyViolation(res.Error) {
			apierr.Conflict(c, "this branch is used by offers: remove it from them first")
			return
		}
		apierr.Internal(c)
		return
	}
	cache.BumpOffersVersion(ctx, h.Redis)
	c.Status(http.StatusNoContent)
}
