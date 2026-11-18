// Package ratelimit is a small fixed-window limiter per client IP, counted in Redis.
// If Redis is unavailable the request is allowed: a cache outage must not take the
// site down.
package ratelimit

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/deltrexgg/supermarket_offer/api/internal/apierr"
)

// Counter adds one to a key and returns the new count. The key expires after ttl.
type Counter interface {
	Incr(ctx context.Context, key string, ttl time.Duration) (int64, error)
}

type redisCounter struct{ rdb *redis.Client }

// NewRedisCounter counts in Redis with an INCR and an EXPIRE in one round trip.
func NewRedisCounter(rdb *redis.Client) Counter { return redisCounter{rdb} }

func (r redisCounter) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	pipe := r.rdb.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// windowKey names the counter for one client, one limit and one time window.
func windowKey(name, ip string, now time.Time, window time.Duration) string {
	start := now.Unix() / int64(window.Seconds())
	return "rl:" + name + ":" + ip + ":" + strconv.FormatInt(start, 10)
}

// retryAfter is how many whole seconds remain in the current window (at least 1).
func retryAfter(now time.Time, window time.Duration) int {
	w := int64(window.Seconds())
	left := w - now.Unix()%w
	if left < 1 {
		left = 1
	}
	return int(left)
}

// Middleware allows limit requests per window for each client IP. name keeps the
// counters of different limits apart. ClientIP honours the trusted-proxy setting
// of the router, so a forged X-Forwarded-For is ignored unless a trusted proxy sent it.
func Middleware(c Counter, name string, limit int, window time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if c == nil || limit <= 0 {
			ctx.Next()
			return
		}
		now := time.Now()
		n, err := c.Incr(ctx.Request.Context(), windowKey(name, ctx.ClientIP(), now, window), window)
		if err != nil {
			log.Printf("ratelimit %s: %v (allowing request)", name, err)
			ctx.Next()
			return
		}
		if n > int64(limit) {
			ctx.Header("Retry-After", strconv.Itoa(retryAfter(now, window)))
			apierr.TooManyRequests(ctx, "too many requests, please slow down")
			return
		}
		ctx.Next()
	}
}
