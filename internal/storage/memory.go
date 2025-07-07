package storage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MemoryStorage provides an in-memory implementation of the Storage interface
// with automatic cleanup of expired items.
type MemoryStorage struct {
	data   map[string]*memoryItem
	mu     sync.RWMutex
	stopCh chan struct{}
}

type memoryItem struct {
	value     string
	expiresAt time.Time
}

// NewMemoryStorage creates a new memory storage instance with automatic cleanup
func NewMemoryStorage() *MemoryStorage {
	ms := &MemoryStorage{
		data:   make(map[string]*memoryItem),
		stopCh: make(chan struct{}),
	}

	// Start cleanup goroutine
	go ms.cleanup()

	return ms
}

func (m *MemoryStorage) Get(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return "", fmt.Errorf("key '%s' not found", key)
	}

	if time.Now().After(item.expiresAt) {
		return "", fmt.Errorf("key '%s' expired", key)
	}

	return item.value, nil
}

func (m *MemoryStorage) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = &memoryItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

func (m *MemoryStorage) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}

func (m *MemoryStorage) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return false, nil
	}

	if time.Now().After(item.expiresAt) {
		return false, nil
	}

	return true, nil
}

// Increment increases a numeric value by the specified amount
func (m *MemoryStorage) Increment(ctx context.Context, key string, by int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.data[key]
	if !exists || time.Now().After(item.expiresAt) {
		// Create new item with default TTL of 1 hour
		m.data[key] = &memoryItem{
			value:     fmt.Sprintf("%d", by),
			expiresAt: time.Now().Add(time.Hour),
		}
		return by, nil
	}

	// Convert current value to int64 and increment
	var currentVal int64
	if _, err := fmt.Sscanf(item.value, "%d", &currentVal); err != nil {
		currentVal = 0
	}

	currentVal += by
	item.value = fmt.Sprintf("%d", currentVal)

	return currentVal, nil
}

func (m *MemoryStorage) Close() error {
	close(m.stopCh)
	return nil
}

func (m *MemoryStorage) cleanup() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupExpired()
		case <-m.stopCh:
			return
		}
	}
}

func (m *MemoryStorage) cleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, item := range m.data {
		if now.After(item.expiresAt) {
			delete(m.data, key)
		}
	}
}

// Rate limit specific methods

// GetTokens retrieves the current token count for a rate limit key
func (m *MemoryStorage) GetTokens(ctx context.Context, key string) (int64, error) {
	value, err := m.Get(ctx, key)
	if err != nil {
		return 0, nil // Return 0 if key doesn't exist
	}

	var tokens int64
	if _, err := fmt.Sscanf(value, "%d", &tokens); err != nil {
		return 0, fmt.Errorf("invalid token value: %w", err)
	}

	return tokens, nil
}

// SetTokens sets the token count with an optional TTL
func (m *MemoryStorage) SetTokens(ctx context.Context, key string, tokens int64, ttl time.Duration) error {
	return m.Set(ctx, key, fmt.Sprintf("%d", tokens), ttl)
}

// DecrementTokens reduces the token count by 1, returns remaining tokens
func (m *MemoryStorage) DecrementTokens(ctx context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.data[key]
	if !exists || time.Now().After(item.expiresAt) {
		return 0, fmt.Errorf("key '%s' not found or expired", key)
	}

	var currentVal int64
	if _, err := fmt.Sscanf(item.value, "%d", &currentVal); err != nil {
		return 0, fmt.Errorf("invalid token value: %w", err)
	}

	if currentVal <= 0 {
		return 0, fmt.Errorf("no tokens available")
	}

	currentVal--
	item.value = fmt.Sprintf("%d", currentVal)

	return currentVal, nil
}

// IncrementTokens increases the token count by the specified amount
func (m *MemoryStorage) IncrementTokens(ctx context.Context, key string, by int64) (int64, error) {
	return m.Increment(ctx, key, by)
}
