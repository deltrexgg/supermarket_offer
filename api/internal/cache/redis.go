package cache

import (
	"context"
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
