package portal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/deltrexgg/supermarket_offer/api/internal/apierr"
	"github.com/deltrexgg/supermarket_offer/api/internal/models"
	"github.com/deltrexgg/supermarket_offer/api/internal/supabase"
)

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	s := strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if s == "" {
		s = "supermarket"
	}
	if len(s) > 120 {
		s = s[:120]
	}
	return s
}

func randomSuffix() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Register creates the Supabase login, then the supermarket (pending) and its
// users row in one transaction. If the transaction fails, the login is removed
// again so the email can be used for another attempt.
func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "body must be JSON")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)

	f := apierr.Fields{}
	if l := len([]rune(req.Name)); l < 2 || l > 150 {
		f.Add("name", "must be 2 to 150 characters")
	}
	if addr, err := mail.ParseAddress(req.Email); err != nil || addr.Address != req.Email || len(req.Email) > 254 {
		f.Add("email", "must be a valid email address")
	}
	if len(req.Password) < 8 || len(req.Password) > 72 {
		f.Add("password", "must be 8 to 72 characters")
	}
	if len(req.Phone) > 30 {
		f.Add("phone", "must be at most 30 characters")
	}
	if !f.Empty() {
		apierr.Validation(c, f)
		return
	}

	ctx := c.Request.Context()

	var taken int64
	if err := h.DB.WithContext(ctx).Model(&models.User{}).Where("email = ?", req.Email).Count(&taken).Error; err != nil {
		apierr.Internal(c)
		return
	}
	if taken > 0 {
		apierr.Conflict(c, "this email is already registered")
		return
	}

	authID, err := h.Admin.CreateUser(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, supabase.ErrEmailExists) {
			apierr.Conflict(c, "this email is already registered")
			return
		}
		log.Printf("register: create supabase user: %v", err)
		apierr.Internal(c)
		return
	}

	var sm models.Supermarket
	txErr := h.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		slug := slugify(req.Name)
		var n int64
		if err := tx.Model(&models.Supermarket{}).Where("slug = ?", slug).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			slug = slug + "-" + randomSuffix()
		}
		sm = models.Supermarket{Name: req.Name, Slug: slug, Status: "pending"}
		if req.Phone != "" {
			p := req.Phone
			sm.Phone = &p
		}
		if err := tx.Create(&sm).Error; err != nil {
			return err
		}
		id := sm.ID
		return tx.Create(&models.User{AuthUserID: authID, Email: req.Email, Role: "supermarket", SupermarketID: &id}).Error
	})
	if txErr != nil {
		// Undo the login, with its own short deadline: the request may be cancelled.
		cctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if derr := h.Admin.DeleteUser(cctx, authID); derr != nil {
			log.Printf("register: cleanup of supabase user %s failed: %v", authID, derr)
		}
		if isUniqueViolation(txErr) {
			apierr.Conflict(c, "this email is already registered")
			return
		}
		log.Printf("register: transaction: %v", txErr)
		apierr.Internal(c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"supermarket": sm,
		"message":     "registered: sign in with your email and password",
	})
}
