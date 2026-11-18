// Package tenant keeps every supermarket inside its own data. The supermarket id
// always comes from the verified user's row, never from a URL or request body.
package tenant

import "gorm.io/gorm"

// Scope limits a query to one supermarket. A nil id matches nothing, so a
// missing id can never turn into "all supermarkets".
func Scope(supermarketID *uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if supermarketID == nil {
			return db.Where("1 = 0")
		}
		return db.Where("supermarket_id = ?", *supermarketID)
	}
}
