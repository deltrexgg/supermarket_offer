// Package auth verifies Supabase Auth access tokens. The browser signs in with
// supabase-js; this API only checks the token (signature via the project's JWKS,
// issuer, expiry) and maps it to a row in our users table.
package auth

import (
	"errors"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"github.com/deltrexgg/supermarket_offer/api/internal/apierr"
	"github.com/deltrexgg/supermarket_offer/api/internal/models"
)

const ctxUserKey = "auth.user"

type Verifier struct {
	kf     keyfunc.Keyfunc
	issuer string
	db     *gorm.DB
}

// NewVerifier loads the project's public keys (cached and refreshed in the background).
func NewVerifier(supabaseURL string, db *gorm.DB) (*Verifier, error) {
	base := strings.TrimRight(supabaseURL, "/")
	if base == "" {
		return nil, errors.New("SUPABASE_URL is not set")
	}
	kf, err := keyfunc.NewDefault([]string{base + "/auth/v1/.well-known/jwks.json"})
	if err != nil {
		return nil, err
	}
	return &Verifier{kf: kf, issuer: base + "/auth/v1", db: db}, nil
}

// Authenticate checks the Bearer token and loads the user. Unknown Supabase users
// (valid token, no users row) are rejected: signing in alone grants nothing.
func (v *Verifier) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		raw, ok := strings.CutPrefix(h, "Bearer ")
		if !ok || raw == "" {
			apierr.Unauthorized(c, "missing bearer token")
			return
		}
		tok, err := jwt.Parse(raw, v.kf.Keyfunc,
			jwt.WithValidMethods([]string{"ES256", "RS256"}),
			jwt.WithIssuer(v.issuer),
			jwt.WithExpirationRequired(),
		)
		if err != nil || !tok.Valid {
			apierr.Unauthorized(c, "invalid or expired token")
			return
		}
		sub, err := tok.Claims.GetSubject()
		if err != nil || sub == "" {
			apierr.Unauthorized(c, "token has no subject")
			return
		}

		var u models.User
		if err := v.db.WithContext(c.Request.Context()).
			Where("auth_user_id = ?", sub).First(&u).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				apierr.Forbidden(c, "no account for this login")
				return
			}
			apierr.Internal(c)
			return
		}
		c.Set(ctxUserKey, &u)
		c.Next()
	}
}

// RequireRole allows only the listed roles. Use after Authenticate.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := CurrentUser(c)
		if u == nil {
			apierr.Unauthorized(c, "not signed in")
			return
		}
		for _, r := range roles {
			if u.Role == r {
				c.Next()
				return
			}
		}
		apierr.Forbidden(c, "your role cannot do this")
	}
}

// CurrentUser returns the verified user, or nil if Authenticate has not run.
func CurrentUser(c *gin.Context) *models.User {
	if v, ok := c.Get(ctxUserKey); ok {
		if u, ok := v.(*models.User); ok {
			return u
		}
	}
	return nil
}
