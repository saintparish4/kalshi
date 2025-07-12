package testing

import (
	"context"
	"time"

	"kalshi/internal/cache"
)

// MockCache implements cache.Cache for testing
type MockCache struct {
	data map[string][]byte
	ttl  map[string]time.Duration
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string][]byte),
		ttl:  make(map[string]time.Duration),
	}
}

func (m *MockCache) Get(ctx context.Context, key string) ([]byte, error) {
	if data, exists := m.data[key]; exists {
		return data, nil
	}
	return nil, cache.ErrCacheMiss
}

func (m *MockCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.data[key] = value
	m.ttl[key] = ttl
	return nil
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	delete(m.ttl, key)
	return nil
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.data[key]
	return exists, nil
}

func (m *MockCache) Close() error {
	return nil
}
