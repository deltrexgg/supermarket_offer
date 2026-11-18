package seed

import (
	"context"
	"fmt"
	"log"

	"gorm.io/gorm"

	"github.com/deltrexgg/supermarket_offer/api/internal/models"
)

// EnsureAdmin makes sure the platform owner login from ADMIN_EMAIL / ADMIN_PASSWORD
// exists in Supabase Auth and in the users table with role admin. An existing admin
// is left alone, so changing ADMIN_PASSWORD later does not reset it.
func EnsureAdmin(ctx context.Context, db *gorm.DB, supabaseURL, secretKey, email, password string) error {
	var existing models.User
	err := db.WithContext(ctx).Where("email = ?", email).First(&existing).Error
	if err == nil {
		if existing.Role != "admin" {
			return fmt.Errorf("ADMIN_EMAIL %s already belongs to a supermarket login; use another email", email)
		}
		log.Printf("admin: %s already exists", email)
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	authID, err := newAdminClient(supabaseURL, secretKey).ensureUser(ctx, email, password)
	if err != nil {
		return err
	}
	if err := db.WithContext(ctx).Create(&models.User{AuthUserID: authID, Email: email, Role: "admin"}).Error; err != nil {
		return err
	}
	log.Printf("admin: created %s", email)
	return nil
}
