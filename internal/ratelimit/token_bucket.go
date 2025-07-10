package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"kalshi/internal/storage"
)

type TokenBucket struct {
	storage      storage.Storage
	capacity     int
	refillRate   int
	refillPeriod time.Duration
	mu           sync.RWMutex // Add mutex for thread safety
}

func NewTokenBucket(storage storage.Storage, capacity, refillRate int, refillPeriod time.Duration) *TokenBucket {
	return &TokenBucket{
		storage:      storage,
		capacity:     capacity,
		refillRate:   refillRate,
		refillPeriod: refillPeriod,
	}
}

func (tb *TokenBucket) Allow(ctx context.Context, key string) (bool, error) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	bucketKey := fmt.Sprintf("bucket:%s", key)
	timestampKey := fmt.Sprintf("timestamp:%s", key)

	// Get current tokens and last refill time
	currentTokens, err := tb.getCurrentTokens(ctx, bucketKey)
	if err != nil {
		return false, fmt.Errorf("failed to get current tokens: %w", err)
	}

	lastRefill, err := tb.getLastRefill(ctx, timestampKey)
	if err != nil {
		lastRefill = now
	}

	// Calculate tokens to add based on elapsed time
	elapsed := now.Sub(lastRefill)
	tokensToAdd := int((elapsed * time.Duration(tb.refillRate)) / tb.refillPeriod)

	// Update token count
	newTokens := currentTokens + tokensToAdd
	if newTokens > tb.capacity {
		newTokens = tb.capacity
	}

	// Check if we can consume a token
	if newTokens <= 0 {
		return false, nil
	}

	// Consume one token
	newTokens--

	// Update storage with better error handling
	if err := tb.setTokens(ctx, bucketKey, newTokens); err != nil {
		return false, fmt.Errorf("failed to set tokens: %w", err)
	}

	if err := tb.setLastRefill(ctx, timestampKey, now); err != nil {
		return false, fmt.Errorf("failed to set last refill: %w", err)
	}

	return true, nil
}

func (tb *TokenBucket) getCurrentTokens(ctx context.Context, key string) (int, error) {
	value, err := tb.storage.Get(ctx, key)
	if err != nil {
		// If key doesn't exist, start with full capacity
		return tb.capacity, nil
	}

	var tokens int
	if _, err := fmt.Sscanf(value, "%d", &tokens); err != nil {
		// If parsing fails, return full capacity
		return tb.capacity, nil
	}

	// Ensure tokens is within valid range
	if tokens < 0 {
		tokens = 0
	}
	if tokens > tb.capacity {
		tokens = tb.capacity
	}

	return tokens, nil
}

func (tb *TokenBucket) setTokens(ctx context.Context, key string, tokens int) error {
	if tokens < 0 {
		tokens = 0
	}
	return tb.storage.Set(ctx, key, fmt.Sprintf("%d", tokens), time.Hour)
}

func (tb *TokenBucket) getLastRefill(ctx context.Context, key string) (time.Time, error) {
	value, err := tb.storage.Get(ctx, key)
	if err != nil {
		return time.Time{}, err
	}

	var timestamp int64
	if _, err := fmt.Sscanf(value, "%d", &timestamp); err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp format: %w", err)
	}

	return time.Unix(timestamp, 0), nil
}

func (tb *TokenBucket) setLastRefill(ctx context.Context, key string, t time.Time) error {
	return tb.storage.Set(ctx, key, fmt.Sprintf("%d", t.Unix()), time.Hour)
}

// Reset clears the token bucket state for a given key
func (tb *TokenBucket) Reset(ctx context.Context, key string) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	bucketKey := fmt.Sprintf("bucket:%s", key)
	timestampKey := fmt.Sprintf("timestamp:%s", key)

	// Delete both bucket and timestamp keys
	if err := tb.storage.Delete(ctx, bucketKey); err != nil {
		return fmt.Errorf("failed to delete bucket key: %w", err)
	}

	if err := tb.storage.Delete(ctx, timestampKey); err != nil {
		return fmt.Errorf("failed to delete timestamp key: %w", err)
	}

	return nil
}
