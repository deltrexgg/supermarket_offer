// Package apierr is the single error format for the whole API:
//
//	{"error": {"code": "validation_failed", "message": "...", "fields": {"offer_price": "must be less than regular_price"}}}
package apierr

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func write(c *gin.Context, status int, code, msg string, fields map[string]string) {
	c.AbortWithStatusJSON(status, gin.H{"error": Body{Code: code, Message: msg, Fields: fields}})
}

func BadRequest(c *gin.Context, msg string) { write(c, http.StatusBadRequest, "bad_request", msg, nil) }
func Unauthorized(c *gin.Context, msg string) {
	write(c, http.StatusUnauthorized, "unauthorized", msg, nil)
}
func Forbidden(c *gin.Context, msg string) { write(c, http.StatusForbidden, "forbidden", msg, nil) }
func NotFound(c *gin.Context, msg string)  { write(c, http.StatusNotFound, "not_found", msg, nil) }
func Conflict(c *gin.Context, msg string)  { write(c, http.StatusConflict, "conflict", msg, nil) }
func Internal(c *gin.Context) {
	write(c, http.StatusInternalServerError, "internal", "something went wrong", nil)
}

// Validation reports per-field problems with HTTP 422.
func Validation(c *gin.Context, fields map[string]string) {
	write(c, http.StatusUnprocessableEntity, "validation_failed", "some fields are invalid", fields)
}

// Fields collects validation problems; call Fail when it is not empty.
type Fields map[string]string

func (f Fields) Add(name, problem string) {
	if _, exists := f[name]; !exists {
		f[name] = problem
	}
}
func (f Fields) Empty() bool { return len(f) == 0 }

func PayloadTooLarge(c *gin.Context, msg string) {
	write(c, http.StatusRequestEntityTooLarge, "payload_too_large", msg, nil)
}
func UnsupportedMedia(c *gin.Context, msg string) {
	write(c, http.StatusUnsupportedMediaType, "unsupported_media_type", msg, nil)
}

func TooManyRequests(c *gin.Context, msg string) {
	write(c, http.StatusTooManyRequests, "rate_limited", msg, nil)
}
