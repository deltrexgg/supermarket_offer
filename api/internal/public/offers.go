package public

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/deltrexgg/supermarket_offer/api/internal/apierr"
)

type listResponse struct {
	Items    []Card `json:"items"`
	Total    int64  `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// ListOffers: GET /offers?q=&area=&category=&supermarket=&sort=&page=&page_size=
func (h *Handler) ListOffers(c *gin.Context) {
	f := apierr.Fields{}
	norm := url.Values{}

	sortKey := c.DefaultQuery("sort", defaultSort)
	if _, ok := sortOrders[sortKey]; !ok {
		f.Add("sort", "must be lowest_price, highest_discount or ending_soon")
		sortKey = defaultSort
	}
	norm.Set("sort", sortKey)

	ids := map[string]uint{}
	for _, name := range []string{"area", "category", "supermarket"} {
		if raw := strings.TrimSpace(c.Query(name)); raw != "" {
			id, ok := positiveInt(raw)
			if !ok {
				f.Add(name, "must be an id")
				continue
			}
			ids[name] = id
			norm.Set(name, strconv.FormatUint(uint64(id), 10))
		}
	}

	term := strings.TrimSpace(c.Query("q"))
	if len([]rune(term)) > 100 {
		f.Add("q", "must be at most 100 characters")
	}
	if term != "" {
		norm.Set("q", strings.ToLower(term))
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 12
	}
	if size > 48 {
		size = 48
	}
	norm.Set("page", strconv.Itoa(page))
	norm.Set("page_size", strconv.Itoa(size))

	if !f.Empty() {
		apierr.Validation(c, f)
		return
	}

	h.respond(c, h.cacheKey(c.Request.Context(), "list", norm), func() ([]byte, error) {
		q := h.DB.WithContext(c.Request.Context()).
			Table("offers_with_status o").
			Joins("JOIN supermarkets s ON s.id = o.supermarket_id AND s.status = 'approved'").
			Where("o.status = 'active'")
		if id, ok := ids["supermarket"]; ok {
			q = q.Where("o.supermarket_id = ?", id)
		}
		if id, ok := ids["category"]; ok {
			q = q.Where("o.category_id = ?", id)
		}
		if id, ok := ids["area"]; ok {
			q = q.Where("EXISTS (SELECT 1 FROM offer_branches ob JOIN branches b ON b.id = ob.branch_id WHERE ob.offer_id = o.id AND b.area_id = ?)", id)
		}
		if term != "" {
			like := "%" + strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(term) + "%"
			q = q.Where("(o.item_name ILIKE ? OR o.brand ILIKE ? OR o.item_name_ar ILIKE ?)", like, like, like)
		}

		q = q.Session(&gorm.Session{})
		var total int64
		if err := q.Count(&total).Error; err != nil {
			return nil, err
		}
		var rows []CardRow
		err := q.Select(`o.id, o.item_name, o.item_name_ar, o.brand, o.unit, o.regular_price, o.offer_price,
			o.discount_pct, to_char(o.valid_to, 'YYYY-MM-DD') AS valid_to, o.days_left,
			o.photo_path, s.id AS supermarket_id, s.name AS supermarket_name, s.logo_path,
			(SELECT count(*) FROM offer_branches ob WHERE ob.offer_id = o.id) AS branch_count`).
			Order(sortOrders[sortKey]).Limit(size).Offset((page - 1) * size).
			Scan(&rows).Error
		if err != nil {
			return nil, err
		}
		items := make([]Card, 0, len(rows))
		for _, r := range rows {
			items = append(items, h.toCard(r))
		}
		return marshal(listResponse{Items: items, Total: total, Page: page, PageSize: size})
	})
}

type branchOut struct {
	ID       uint    `json:"id"`
	Name     string  `json:"name"`
	Address  string  `json:"address"`
	Phone    *string `json:"phone"`
	AreaID   uint    `json:"area_id"`
	AreaName string  `json:"area_name"`
}

type detailRow struct {
	CardRow
	ValidFrom     string
	Description   *string
	DescriptionAr *string
	CategoryID    uint
	CategoryName  string
	ShopPhone     *string
}

type detailResponse struct {
	Card
	ValidFrom     string      `json:"valid_from"`
	Description   *string     `json:"description"`
	DescriptionAr *string     `json:"description_ar"`
	CategoryID    uint        `json:"category_id"`
	CategoryName  string      `json:"category_name"`
	ShopPhone     *string     `json:"supermarket_phone"`
	Branches      []branchOut `json:"branches"`
}

// GetOffer: GET /offers/:id. 404 when the offer is not active today or its supermarket is not approved.
func (h *Handler) GetOffer(c *gin.Context) {
	id, ok := positiveInt(c.Param("id"))
	if !ok {
		apierr.NotFound(c, "offer not found")
		return
	}
	norm := url.Values{"id": {strconv.FormatUint(uint64(id), 10)}}

	h.respond(c, h.cacheKey(c.Request.Context(), "detail", norm), func() ([]byte, error) {
		ctx := c.Request.Context()
		var rows []detailRow
		err := h.DB.WithContext(ctx).Table("offers_with_status o").
			Joins("JOIN supermarkets s ON s.id = o.supermarket_id AND s.status = 'approved'").
			Joins("JOIN categories cat ON cat.id = o.category_id").
			Where("o.id = ? AND o.status = 'active'", id).
			Select(`o.id, o.item_name, o.item_name_ar, o.brand, o.unit, o.regular_price, o.offer_price,
				o.discount_pct, to_char(o.valid_to, 'YYYY-MM-DD') AS valid_to,
				to_char(o.valid_from, 'YYYY-MM-DD') AS valid_from, o.days_left, o.photo_path,
				o.description, o.description_ar, o.category_id, cat.name AS category_name,
				s.id AS supermarket_id, s.name AS supermarket_name, s.logo_path, s.phone AS shop_phone,
				(SELECT count(*) FROM offer_branches ob WHERE ob.offer_id = o.id) AS branch_count`).
			Scan(&rows).Error
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return nil, errNotFound
		}
		r := rows[0]

		var branches []branchOut
		if err := h.DB.WithContext(ctx).Table("offer_branches ob").
			Joins("JOIN branches b ON b.id = ob.branch_id").
			Joins("JOIN areas a ON a.id = b.area_id").
			Where("ob.offer_id = ?", id).
			Select("b.id, b.name, b.address, b.phone, a.id AS area_id, a.name AS area_name").
			Order("a.name, b.name").Scan(&branches).Error; err != nil {
			return nil, err
		}
		if branches == nil {
			branches = []branchOut{}
		}
		return marshal(detailResponse{
			Card: h.toCard(r.CardRow), ValidFrom: r.ValidFrom,
			Description: r.Description, DescriptionAr: r.DescriptionAr,
			CategoryID: r.CategoryID, CategoryName: r.CategoryName, ShopPhone: r.ShopPhone,
			Branches: branches,
		})
	})
}

type shopOut struct {
	ID      uint    `json:"id"`
	Name    string  `json:"name"`
	Slug    string  `json:"slug"`
	LogoURL *string `json:"logo_url"`
}

// ListSupermarkets: GET /supermarkets, the approved ones, for the filter.
func (h *Handler) ListSupermarkets(c *gin.Context) {
	h.respond(c, h.cacheKey(c.Request.Context(), "shops", url.Values{}), func() ([]byte, error) {
		var rows []struct {
			ID       uint
			Name     string
			Slug     string
			LogoPath *string
		}
		if err := h.DB.WithContext(c.Request.Context()).Table("supermarkets").
			Select("id, name, slug, logo_path").Where("status = 'approved'").Order("name").
			Scan(&rows).Error; err != nil {
			return nil, err
		}
		items := make([]shopOut, 0, len(rows))
		for _, r := range rows {
			items = append(items, shopOut{ID: r.ID, Name: r.Name, Slug: r.Slug, LogoURL: h.fileURL(r.LogoPath)})
		}
		return marshal(gin.H{"items": items})
	})
}
