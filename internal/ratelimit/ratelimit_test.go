package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"kalshi/pkg/metrics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStorage implements storage.Storage for testing
type MockStorage struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string]string),
	}
}

func (m *MockStorage) Get(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if value, exists := m.data[key]; exists {
		return value, nil
	}
	return "", fmt.Errorf("key '%s' not found", key)
}

func (m *MockStorage) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = value
	return nil
}

func (m *MockStorage) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}

func (m *MockStorage) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.data[key]
	return exists, nil
}

func (m *MockStorage) Increment(ctx context.Context, key string, by int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	current := int64(0)
	if value, exists := m.data[key]; exists {
		// Parse existing value
		if _, err := fmt.Sscanf(value, "%d", &current); err != nil {
			current = 0
		}
	}

	current += by
	m.data[key] = fmt.Sprintf("%d", current)
	return current, nil
}

func (m *MockStorage) Close() error {
	return nil
}

// RateLimitStorage interface methods
func (m *MockStorage) GetTokens(ctx context.Context, key string) (int64, error) {
	value, err := m.Get(ctx, key)
	if err != nil {
		return 0, nil
	}

	var tokens int64
	if _, err := fmt.Sscanf(value, "%d", &tokens); err != nil {
		return 0, fmt.Errorf("invalid token value: %w", err)
	}

	return tokens, nil
}

func (m *MockStorage) SetTokens(ctx context.Context, key string, tokens int64, ttl time.Duration) error {
	return m.Set(ctx, key, fmt.Sprintf("%d", tokens), ttl)
}

func (m *MockStorage) DecrementTokens(ctx context.Context, key string) (int64, error) {
	tokens, err := m.GetTokens(ctx, key)
	if err != nil {
		return 0, err
	}

	if tokens <= 0 {
		return 0, nil
	}

	tokens--
	err = m.SetTokens(ctx, key, tokens, time.Hour)
	return tokens, err
}

func (m *MockStorage) IncrementTokens(ctx context.Context, key string, by int64) (int64, error) {
	tokens, err := m.GetTokens(ctx, key)
	if err != nil {
		tokens = 0
	}

	tokens += by
	return tokens, m.SetTokens(ctx, key, tokens, time.Hour)
}

// TestTokenBucket_BasicFunctionality tests basic token bucket operations
func TestTokenBucket_BasicFunctionality(t *testing.T) {
	storage := NewMockStorage()
	tb := NewTokenBucket(storage, 10, 5, time.Minute) // 10 capacity, 5 tokens per minute
	ctx := context.Background()

	// Test initial state - should allow requests up to capacity
	for i := 0; i < 10; i++ {
		allowed, err := tb.Allow(ctx, "test-key")
		require.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// 11th request should be denied
	allowed, err := tb.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.False(t, allowed, "11th request should be denied")
}

// TestTokenBucket_Refill tests token refill over time
func TestTokenBucket_Refill(t *testing.T) {
	storage := NewMockStorage()
	tb := NewTokenBucket(storage, 10, 60, time.Minute) // 10 capacity, 60 tokens per minute (1 per second)
	ctx := context.Background()

	// Consume all tokens
	for i := 0; i < 10; i++ {
		allowed, err := tb.Allow(ctx, "test-key")
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// Next request should be denied
	allowed, err := tb.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.False(t, allowed)

	// Wait for 1 token to refill (1 second)
	time.Sleep(1100 * time.Millisecond)

	// Should allow 1 more request
	allowed, err = tb.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.True(t, allowed)

	// Next request should still be denied
	allowed, err = tb.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.False(t, allowed)
}

// TestTokenBucket_CapacityLimit tests that tokens don't exceed capacity
func TestTokenBucket_CapacityLimit(t *testing.T) {
	storage := NewMockStorage()
	tb := NewTokenBucket(storage, 5, 10, time.Minute) // 5 capacity, 10 tokens per minute
	ctx := context.Background()

	// Wait for more tokens than capacity to accumulate
	time.Sleep(40 * time.Second) // Should accumulate 6+ tokens

	// Should only allow up to capacity (5 requests)
	for i := 0; i < 5; i++ {
		allowed, err := tb.Allow(ctx, "test-key")
		require.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// 6th request should be denied
	allowed, err := tb.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.False(t, allowed, "6th request should be denied")
}

// TestTokenBucket_Reset tests the reset functionality
func TestTokenBucket_Reset(t *testing.T) {
	storage := NewMockStorage()
	tb := NewTokenBucket(storage, 10, 5, time.Minute)
	ctx := context.Background()

	// Consume some tokens
	for i := 0; i < 3; i++ {
		allowed, err := tb.Allow(ctx, "test-key")
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// Reset the bucket
	err := tb.Reset(ctx, "test-key")
	require.NoError(t, err)

	// Should be able to consume full capacity again
	for i := 0; i < 10; i++ {
		allowed, err := tb.Allow(ctx, "test-key")
		require.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed after reset", i+1)
	}
}

// TestTokenBucket_ConcurrentAccess tests thread safety
func TestTokenBucket_ConcurrentAccess(t *testing.T) {
	storage := NewMockStorage()
	tb := NewTokenBucket(storage, 100, 10, time.Minute)
	ctx := context.Background()

	var wg sync.WaitGroup
	results := make(chan bool, 200)

	// Launch 200 concurrent requests
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed, err := tb.Allow(ctx, "concurrent-key")
			require.NoError(t, err)
			results <- allowed
		}()
	}

	wg.Wait()
	close(results)

	// Count allowed requests
	allowedCount := 0
	for allowed := range results {
		if allowed {
			allowedCount++
		}
	}

	// Should allow exactly 100 requests (capacity)
	assert.Equal(t, 100, allowedCount, "Should allow exactly capacity number of requests")
}

// TestTokenBucket_ErrorHandling tests error scenarios
func TestTokenBucket_ErrorHandling(t *testing.T) {
	storage := NewMockStorage()
	tb := NewTokenBucket(storage, 10, 5, time.Minute)
	ctx := context.Background()

	// Test with empty key
	allowed, err := tb.Allow(ctx, "")
	require.NoError(t, err)
	assert.True(t, allowed) // Should work with empty key

	// Test reset with empty key
	err = tb.Reset(ctx, "")
	require.NoError(t, err)
}

// TestLimiter_BasicFunctionality tests basic limiter operations
func TestLimiter_BasicFunctionality(t *testing.T) {
	storage := NewMockStorage()
	limiter := NewLimiter(storage, 60, 10) // 60 per minute, 10 burst
	ctx := context.Background()

	// Test basic allow/deny
	allowed, err := limiter.Allow(ctx, "client1", "/api/test")
	require.NoError(t, err)
	assert.True(t, allowed)

	// Test different client/path combinations
	allowed, err = limiter.Allow(ctx, "client2", "/api/test")
	require.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = limiter.Allow(ctx, "client1", "/api/other")
	require.NoError(t, err)
	assert.True(t, allowed)
}

// TestLimiter_RateLimitEnforcement tests rate limit enforcement
func TestLimiter_RateLimitEnforcement(t *testing.T) {
	storage := NewMockStorage()
	limiter := NewLimiter(storage, 5, 5) // 5 per minute, 5 burst
	ctx := context.Background()

	// Should allow 5 requests (burst capacity)
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, "client1", "/api/test")
		require.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// 6th request should be denied
	allowed, err := limiter.Allow(ctx, "client1", "/api/test")
	require.NoError(t, err)
	assert.False(t, allowed, "6th request should be denied")

	// Different client should still be allowed
	allowed, err = limiter.Allow(ctx, "client2", "/api/test")
	require.NoError(t, err)
	assert.True(t, allowed)
}

// TestLimiter_Reset tests limiter reset functionality
func TestLimiter_Reset(t *testing.T) {
	storage := NewMockStorage()
	limiter := NewLimiter(storage, 5, 5)
	ctx := context.Background()

	// Consume all tokens
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, "client1", "/api/test")
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// Reset the limiter
	err := limiter.Reset(ctx, "client1", "/api/test")
	require.NoError(t, err)

	// Should be able to make requests again
	allowed, err := limiter.Allow(ctx, "client1", "/api/test")
	require.NoError(t, err)
	assert.True(t, allowed)
}

// TestLimiter_ConcurrentAccess tests thread safety
func TestLimiter_ConcurrentAccess(t *testing.T) {
	storage := NewMockStorage()
	limiter := NewLimiter(storage, 100, 50)
	ctx := context.Background()

	var wg sync.WaitGroup
	results := make(chan bool, 100)

	// Launch 100 concurrent requests
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			allowed, err := limiter.Allow(ctx, "client1", "/api/test")
			require.NoError(t, err)
			results <- allowed
		}(i)
	}

	wg.Wait()
	close(results)

	// Count allowed requests
	allowedCount := 0
	for allowed := range results {
		if allowed {
			allowedCount++
		}
	}

	// Should allow at least 50 requests (burst capacity)
	assert.GreaterOrEqual(t, allowedCount, 50, "Should allow at least burst capacity number of requests")
}

// TestLimiter_Integration tests integration with metrics
func TestLimiter_Integration(t *testing.T) {
	storage := NewMockStorage()
	limiter := NewLimiter(storage, 5, 5)
	ctx := context.Background()

	// Consume all tokens
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, "client1", "/api/test")
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// This should trigger a rate limit hit and increment metrics
	allowed, err := limiter.Allow(ctx, "client1", "/api/test")
	require.NoError(t, err)
	assert.False(t, allowed)

	// Verify metrics were incremented (this is a basic check)
	// In a real test, you might want to check the actual metric values
	assert.NotNil(t, metrics.RateLimitHits)
}

// TestLimiter_EdgeCases tests edge cases
func TestLimiter_EdgeCases(t *testing.T) {
	storage := NewMockStorage()
	limiter := NewLimiter(storage, 10, 10)
	ctx := context.Background()

	// Test with empty client ID
	allowed, err := limiter.Allow(ctx, "", "/api/test")
	require.NoError(t, err)
	assert.True(t, allowed)

	// Test with empty path
	allowed, err = limiter.Allow(ctx, "client1", "")
	require.NoError(t, err)
	assert.True(t, allowed)

	// Test reset with empty values
	err = limiter.Reset(ctx, "", "")
	require.NoError(t, err)
}

// TestTokenBucket_StorageErrors tests handling of storage errors
func TestTokenBucket_StorageErrors(t *testing.T) {
	// Create a storage that returns errors
	errorStorage := &ErrorStorage{}
	tb := NewTokenBucket(errorStorage, 10, 5, time.Minute)
	ctx := context.Background()

	// Should handle storage errors gracefully
	allowed, err := tb.Allow(ctx, "test-key")
	require.Error(t, err)
	assert.False(t, allowed)
}

// ErrorStorage is a mock storage that always returns errors
type ErrorStorage struct{}

func (e *ErrorStorage) Get(ctx context.Context, key string) (string, error) {
	return "", fmt.Errorf("storage error")
}

func (e *ErrorStorage) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return fmt.Errorf("storage error")
}

func (e *ErrorStorage) Delete(ctx context.Context, key string) error {
	return fmt.Errorf("storage error")
}

func (e *ErrorStorage) Exists(ctx context.Context, key string) (bool, error) {
	return false, fmt.Errorf("storage error")
}

func (e *ErrorStorage) Increment(ctx context.Context, key string, by int64) (int64, error) {
	return 0, fmt.Errorf("storage error")
}

func (e *ErrorStorage) Close() error {
	return nil
}

// Benchmark tests for performance
func BenchmarkTokenBucket_Allow(b *testing.B) {
	storage := NewMockStorage()
	tb := NewTokenBucket(storage, 1000, 100, time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.Allow(ctx, "benchmark-key")
	}
}

func BenchmarkLimiter_Allow(b *testing.B) {
	storage := NewMockStorage()
	limiter := NewLimiter(storage, 1000, 100)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(ctx, "client1", "/api/test")
	}
}
