package ratelimit

import (
	"context"
	"fmt"
	"time"

	"kalshi/internal/storage"
	"kalshi/pkg/metrics"
)

type Limiter struct {
	storage       storage.Storage
	defaultRate   int
	burstCapacity int
	tokenBucket   *TokenBucket
}

func NewLimiter(storage storage.Storage, defaultRate, burstCapacity int) *Limiter {
	tb := NewTokenBucket(storage, burstCapacity, defaultRate, time.Minute)

	return &Limiter{
		storage:       storage,
		defaultRate:   defaultRate,
		burstCapacity: burstCapacity,
		tokenBucket:   tb,
	}
}

func (l *Limiter) Allow(ctx context.Context, clientID, path string) (bool, error) {
	key := fmt.Sprintf("ratelimit:%s:%s", clientID, path)

	// Use token bucket for rate limiting
	allowed, err := l.tokenBucket.Allow(ctx, key)
	if err != nil {
		return false, err
	}

	if !allowed {
		// Record rate limit hit
		metrics.RateLimitHits.WithLabelValues(clientID, path).Inc()
	}

	return allowed, nil
}

func (l *Limiter) Reset(ctx context.Context, clientID, path string) error {
	key := fmt.Sprintf("ratelimit:%s:%s", clientID, path)
	return l.tokenBucket.Reset(ctx, key)
}
