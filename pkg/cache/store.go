package cache

import (
	"context"
	"time"
)

// Store represents a cache backend
type Store interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key string, value string, ttl time.Duration) error
	Forget(ctx context.Context, key string) error
	Flush(ctx context.Context) error
}
