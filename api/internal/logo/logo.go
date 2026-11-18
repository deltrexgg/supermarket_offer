// Package logo holds the pieces the portal and the admin area share for
// supermarket logos: where they are stored, and how an upload is validated.
package logo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/deltrexgg/supermarket_offer/api/internal/apierr"
	"github.com/deltrexgg/supermarket_offer/api/internal/storage"
)

// Prefix is the storage folder for supermarket logos. Keys this app writes look
// like supermarket-logo/7/<uuid>.png. A file placed in that folder by hand under
// any other name works just as well, and is never deleted by this app.
const Prefix = "supermarket-logo"

// Key is where a newly uploaded logo is stored.
func Key(supermarketID uint, ext string) string {
	return fmt.Sprintf("%s/%d/%s%s", Prefix, supermarketID, uuid.NewString(), ext)
}

// Owned reports whether a stored key was written by this app for this
// supermarket. Only those are deleted when a logo is replaced or removed.
func Owned(supermarketID uint, key string) bool {
	return strings.HasPrefix(key, fmt.Sprintf("%s/%d/", Prefix, supermarketID))
}

// Read validates the "logo" file field. It returns (nil, true) when the request
// carries no logo, since a logo is always optional, and answers the request
// itself when a file was sent but cannot be used.
func Read(c *gin.Context) (*storage.Image, bool) {
	fh, err := c.FormFile("logo")
	if err != nil {
		return nil, true
	}
	img, err := storage.ReadImage(fh)
	if err == nil {
		return img, true
	}
	switch {
	case errors.Is(err, storage.ErrTooLarge):
		apierr.PayloadTooLarge(c, "the logo must be 5 MB or smaller")
	case errors.Is(err, storage.ErrUnsupported):
		apierr.UnsupportedMedia(c, "the logo must be a JPEG, PNG, WebP, GIF or AVIF image")
	default:
		apierr.BadRequest(c, "could not read the uploaded logo")
	}
	return nil, false
}

// RemoveFile deletes a stored key on its own deadline, so a cancelled request
// does not leave the file behind.
func RemoveFile(store storage.Store, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := store.Delete(ctx, key); err != nil {
		log.Printf("storage cleanup %s: %v", key, err)
	}
}
