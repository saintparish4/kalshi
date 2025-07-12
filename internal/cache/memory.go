package cache

import (
	"context"
	"sync"
	"time"
)

// MemoryCache provides an in-memory cache implementation
type MemoryCache struct {
	data    map[string]cacheItem
	mutex   sync.RWMutex
	maxSize int
	ttl     time.Duration
}

// cacheItem represents a cached item with expiration
type cacheItem struct {
	value      []byte
	expiration time.Time
}

// NewMemoryCache creates a new memory cache with specified max size and TTL
func NewMemoryCache(maxSize int, ttl time.Duration) *MemoryCache {
	cache := &MemoryCache{
		data:    make(map[string]cacheItem),
		maxSize: maxSize,
		ttl:     ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a value from memory cache
func (mc *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	item, exists := mc.data[key]
	if !exists {
		return nil, ErrCacheMiss
	}

	// Check if item has expired
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		// Remove expired item
		mc.mutex.RUnlock()
		mc.mutex.Lock()
		delete(mc.data, key)
		mc.mutex.Unlock()
		mc.mutex.RLock()
		return nil, ErrCacheMiss
	}

	return item.value, nil
}

// Set stores a value in memory cache
func (mc *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// If max size reached, remove oldest item
	if len(mc.data) >= mc.maxSize {
		mc.evictOldest()
	}

	// Calculate expiration
	expiration := time.Time{}
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	} else if mc.ttl > 0 {
		expiration = time.Now().Add(mc.ttl)
	}

	mc.data[key] = cacheItem{
		value:      value,
		expiration: expiration,
	}

	return nil
}

// Delete removes a key from memory cache
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	delete(mc.data, key)
	return nil
}

// Exists checks if a key exists in memory cache
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	item, exists := mc.data[key]
	if !exists {
		return false, nil
	}

	// Check if item has expired
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		return false, nil
	}

	return true, nil
}

// Close releases resources (no-op for memory cache)
func (mc *MemoryCache) Close() error {
	return nil
}

// evictOldest removes the oldest item from cache
func (mc *MemoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range mc.data {
		if oldestKey == "" || item.expiration.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.expiration
		}
	}

	if oldestKey != "" {
		delete(mc.data, oldestKey)
	}
}

// cleanup periodically removes expired items
func (mc *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		mc.mutex.Lock()
		now := time.Now()
		for key, item := range mc.data {
			if !item.expiration.IsZero() && now.After(item.expiration) {
				delete(mc.data, key)
			}
		}
		mc.mutex.Unlock()
	}
}
