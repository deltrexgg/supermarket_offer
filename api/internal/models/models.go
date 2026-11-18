package models

import (
	"time"

	"gorm.io/gorm"
)

// Merchant is a supermarket that publishes offers.
type Merchant struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:200;not null" json:"name"`
	Slug      string         `gorm:"size:200;uniqueIndex;not null" json:"slug"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Offer is a single deal published by a merchant.
type Offer struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	MerchantID uint           `gorm:"index;not null" json:"merchant_id"`
	Title      string         `gorm:"size:255;not null" json:"title"`
	StartsAt   *time.Time     `json:"starts_at"`
	EndsAt     *time.Time     `json:"ends_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func All() []any {
	return []any{&Merchant{}, &Offer{}}
}
