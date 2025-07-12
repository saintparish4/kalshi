package cache

import (
	"context"
	"time"
)

// Manager provides a two-tier cache system with L1 (memory) and L2 (Redis) caches
type Manager struct {
	l1Cache Cache // Memory cache (fast)
	l2Cache Cache // Redis cache (persistent)
	useL2   bool  // Whether to use L2 cache
}

// NewManager creates a new cache manager with L1 and optional L2 cache
func NewManager(l1Cache Cache, l2Cache Cache, useL2 bool) *Manager {
	return &Manager{
		l1Cache: l1Cache,
		l2Cache: l2Cache,
		useL2:   useL2 && l2Cache != nil,
	}
}

// Get retrieves a value from cache, checking L1 first, then L2
func (m *Manager) Get(ctx context.Context, key string) ([]byte, error) {
	// Try L1 cache first
	value, err := m.l1Cache.Get(ctx, key)
	if err == nil {
		return value, nil
	}

	// If L1 miss and L2 is available, try L2
	if m.useL2 {
		value, err = m.l2Cache.Get(ctx, key)
		if err == nil {
			// Populate L1 cache with the value from L2
			_ = m.l1Cache.Set(ctx, key, value, 0) // Use default TTL
			return value, nil
		}
	}

	return nil, ErrCacheMiss
}

// Set stores a value in both L1 and L2 caches
func (m *Manager) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	// Set in L1 cache
	if err := m.l1Cache.Set(ctx, key, value, ttl); err != nil {
		return err
	}

	// Set in L2 cache if available
	if m.useL2 {
		if err := m.l2Cache.Set(ctx, key, value, ttl); err != nil {
			// Log error but don't fail the operation
			// L1 cache is still available
		}
	}

	return nil
}

// Delete removes a key from both L1 and L2 caches
func (m *Manager) Delete(ctx context.Context, key string) error {
	// Delete from L1
	if err := m.l1Cache.Delete(ctx, key); err != nil {
		return err
	}

	// Delete from L2 if available
	if m.useL2 {
		_ = m.l2Cache.Delete(ctx, key) // Ignore errors for L2
	}

	return nil
}

// Exists checks if a key exists in either cache
func (m *Manager) Exists(ctx context.Context, key string) (bool, error) {
	// Check L1 first
	exists, err := m.l1Cache.Exists(ctx, key)
	if err == nil && exists {
		return true, nil
	}

	// Check L2 if available
	if m.useL2 {
		exists, err = m.l2Cache.Exists(ctx, key)
		if err == nil && exists {
			return true, nil
		}
	}

	return false, nil
}

// Close closes both L1 and L2 caches
func (m *Manager) Close() error {
	var l1Err, l2Err error

	// Close L1 cache
	if m.l1Cache != nil {
		l1Err = m.l1Cache.Close()
	}

	// Close L2 cache if available
	if m.useL2 && m.l2Cache != nil {
		l2Err = m.l2Cache.Close()
	}

	// Return first error if any
	if l1Err != nil {
		return l1Err
	}
	return l2Err
}
