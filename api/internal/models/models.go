// Package models maps the tables created by api/migrations. The schema is owned
// by the SQL migrations, so GORM never creates or alters tables.
package models

import "time"

type Area struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`
}

type Category struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`
}

type Supermarket struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	LogoPath  *string   `json:"logo_path"`
	Phone     *string   `json:"phone"`
	Status    string    `json:"status"` // pending | approved | blocked
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	Role          string    `json:"role"` // admin | supermarket
	SupermarketID *uint     `json:"supermarket_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type Branch struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	SupermarketID uint      `json:"supermarket_id"`
	AreaID        uint      `json:"area_id"`
	Name          string    `json:"name"`
	Address       string    `json:"address"`
	Phone         *string   `json:"phone"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Offer prices are AED with 2 decimals. DiscountPct is computed by the database
// (generated column), so it is read-only here: GORM never writes it.
type Offer struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	SupermarketID uint      `json:"supermarket_id"`
	CategoryID    uint      `json:"category_id"`
	ItemName      string    `json:"item_name"`
	ItemNameAr    *string   `json:"item_name_ar"`
	Brand         *string   `json:"brand"`
	Unit          string    `json:"unit"`
	Description   *string   `json:"description"`
	DescriptionAr *string   `json:"description_ar"`
	RegularPrice  float64   `json:"regular_price"`
	OfferPrice    float64   `json:"offer_price"`
	DiscountPct   int16     `gorm:"->" json:"discount_pct"`
	ValidFrom     string    `json:"valid_from"` // YYYY-MM-DD, Dubai calendar day
	ValidTo       string    `json:"valid_to"`
	PhotoPath     *string   `json:"photo_path"`
	PosterPath    *string   `json:"poster_path"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type OfferBranch struct {
	OfferID       uint `gorm:"primaryKey"`
	BranchID      uint `gorm:"primaryKey"`
	SupermarketID uint
}

// OfferWithStatus reads the offers_with_status view (read-only).
// Status is upcoming | active | expired, computed from today's date in Dubai.
// DaysLeft is 0 on the last valid day ("Ends today").
type OfferWithStatus struct {
	Offer
	Status   string `json:"status"`
	DaysLeft int    `json:"days_left"`
}

func (OfferWithStatus) TableName() string { return "offers_with_status" }
