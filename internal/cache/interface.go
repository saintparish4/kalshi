package cache

import (
	"context"
	"time"
)

// Cache defines the interface for cache operations
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Close() error
}

// Ensure RedisCache implements the Cache interface
var _ Cache = (*RedisCache)(nil)
