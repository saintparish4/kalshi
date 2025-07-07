package storage

import (
	"context"
	"time"
)

// Storage provides a generic key-value storage interface for basic operations
type Storage interface {
	// Get retrieves a value by key, returning an error if not found
	Get(ctx context.Context, key string) (string, error)
	// Set stores a value with an optional TTL (time-to-live)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	// Delete removes a key-value pair
	Delete(ctx context.Context, key string) error
	// Exists checks if a key exists in storage
	Exists(ctx context.Context, key string) (bool, error)
	// Increment increases a numeric value by the specified amount
	Increment(ctx context.Context, key string, by int64) (int64, error)
	// Close releases any resources held by the storage implementation
	Close() error
}

// RateLimitStorage provides specialized storage operations for rate limiting
type RateLimitStorage interface {
	// GetTokens retrieves the current token count for a rate limit key
	GetTokens(ctx context.Context, key string) (int64, error)
	// SetTokens sets the token count with an optional TTL
	SetTokens(ctx context.Context, key string, tokens int64, ttl time.Duration) error
	// DecrementTokens reduces the token count by 1, returns remaining tokens
	DecrementTokens(ctx context.Context, key string) (int64, error)
	// IncrementTokens increases the token count by the specified amount
	IncrementTokens(ctx context.Context, key string, by int64) (int64, error)
}
