package cache

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func Connect(addr, password string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr, Password: password})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}

// OffersVersionKey is bumped after every offer, branch or supermarket change.
// Public list cache keys include it, so a bump invalidates them all at once.
const OffersVersionKey = "offers:version"

// BumpOffersVersion is called after a database commit. A Redis failure is logged
// but never fails the request: the cache TTL ends at Dubai midnight anyway.
func BumpOffersVersion(ctx context.Context, rdb *redis.Client) {
	if rdb == nil {
		return
	}
	if err := rdb.Incr(ctx, OffersVersionKey).Err(); err != nil {
		log.Printf("cache: bump %s: %v", OffersVersionKey, err)
	}
}

// Remember returns the cached bytes for key, or builds them, stores them with the
// given TTL and returns them. The bool is true on a cache hit. A Redis problem is
// logged and the data is built from the database instead: the cache never breaks a request.
func Remember(ctx context.Context, rdb *redis.Client, key string, ttl time.Duration, build func() ([]byte, error)) ([]byte, bool, error) {
	if rdb != nil {
		switch b, err := rdb.Get(ctx, key).Bytes(); {
		case err == nil:
			return b, true, nil
		case !errors.Is(err, redis.Nil):
			log.Printf("cache: get %s: %v", key, err)
		}
	}
	b, err := build()
	if err != nil {
		return nil, false, err
	}
	if rdb != nil && ttl > 0 {
		if err := rdb.Set(ctx, key, b, ttl).Err(); err != nil {
			log.Printf("cache: set %s: %v", key, err)
		}
	}
	return b, false, nil
}

// OffersVersion reads the current version number ("0" when none was set yet).
func OffersVersion(ctx context.Context, rdb *redis.Client) string {
	if rdb == nil {
		return "0"
	}
	v, err := rdb.Get(ctx, OffersVersionKey).Result()
	if err != nil {
		return "0"
	}
	return v
}
