package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kalshi/internal/ratelimit"
	"kalshi/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock rate limit storage for testing
type MockRateLimitStorage struct {
	data map[string]string
}

func NewMockRateLimitStorage() *MockRateLimitStorage {
	return &MockRateLimitStorage{
		data: make(map[string]string),
	}
}

func (m *MockRateLimitStorage) Get(ctx context.Context, key string) (string, error) {
	if value, exists := m.data[key]; exists {
		return value, nil
	}
	return "", assert.AnError
}

func (m *MockRateLimitStorage) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *MockRateLimitStorage) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockRateLimitStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.data[key]
	return exists, nil
}

func (m *MockRateLimitStorage) Increment(ctx context.Context, key string, by int64) (int64, error) {
	// Simple increment implementation for testing
	return 1, nil
}

func (m *MockRateLimitStorage) Close() error {
	return nil
}

func TestRateLimit(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	// Create a real limiter with mock storage
	mockStorage := NewMockRateLimitStorage()
	limiter := ratelimit.NewLimiter(mockStorage, 60, 10)

	tests := []struct {
		name           string
		clientID       string
		path           string
		userID         string
		rateLimit      int
		expectedStatus int
		expectedBody   string
		checkHeaders   bool
	}{
		{
			name:           "Rate limit allowed",
			clientID:       "client123",
			path:           "/api/test",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"success"}`,
		},
		{
			name:           "With custom rate limit from context",
			clientID:       "client101",
			path:           "/api/test",
			userID:         "user123",
			rateLimit:      100,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(RateLimit(limiter, log))
			router.GET("/api/test", func(c *gin.Context) {
				// Set custom rate limit if provided
				if tt.rateLimit > 0 {
					c.Set("rate_limit", tt.rateLimit)
				}
				// Set user ID if provided
				if tt.userID != "" {
					c.Set("user_id", tt.userID)
				}
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req := httptest.NewRequest("GET", "/api/test", nil)
			req.RemoteAddr = "127.0.0.1:12345"
			if tt.clientID != "" {
				req.Header.Set(ClientIDHeader, tt.clientID)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			if tt.checkHeaders {
				// Check rate limit headers
				assert.NotEmpty(t, w.Header().Get(RateLimitHeader))
				assert.Equal(t, "0", w.Header().Get(RateLimitRemainingHeader))
				assert.NotEmpty(t, w.Header().Get(RateLimitResetHeader))
			}
		})
	}
}

func TestGetClientID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		clientIDHeader string
		remoteAddr     string
		expectedID     string
	}{
		{
			name:       "User ID from context",
			userID:     "user123",
			remoteAddr: "127.0.0.1:12345",
			expectedID: "user123",
		},
		{
			name:           "Client ID from header",
			clientIDHeader: "client456",
			remoteAddr:     "127.0.0.1:12345",
			expectedID:     "client456",
		},
		{
			name:       "Anonymous user falls back to IP",
			userID:     "anonymous",
			remoteAddr: "127.0.0.1:12345",
			expectedID: "127.0.0.1",
		},
		{
			name:       "No user ID or client ID falls back to IP",
			remoteAddr: "192.168.1.1:54321",
			expectedID: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/test", func(c *gin.Context) {
				if tt.userID != "" {
					c.Set("user_id", tt.userID)
				}
				clientID := getClientID(c)
				c.JSON(http.StatusOK, gin.H{"client_id": clientID})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.clientIDHeader != "" {
				req.Header.Set(ClientIDHeader, tt.clientIDHeader)
			}
			req.RemoteAddr = tt.remoteAddr
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.JSONEq(t, `{"client_id":"`+tt.expectedID+`"}`, w.Body.String())
		})
	}
}

func TestRateLimitHeaders(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	mockStorage := NewMockRateLimitStorage()
	limiter := ratelimit.NewLimiter(mockStorage, 60, 10)

	router := gin.New()
	// Add middleware to set rate limit before rate limiting
	router.Use(func(c *gin.Context) {
		c.Set("rate_limit", 150)
		c.Next()
	})
	router.Use(RateLimit(limiter, log))
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set(ClientIDHeader, "test-client")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"success"}`, w.Body.String())

	// Check that custom rate limit header is set
	assert.Equal(t, "150", w.Header().Get(RateLimitHeader))
}

func TestRateLimitDifferentPaths(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	mockStorage := NewMockRateLimitStorage()
	limiter := ratelimit.NewLimiter(mockStorage, 1, 1) // Very low rate limit for testing

	router := gin.New()
	router.Use(RateLimit(limiter, log))
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/api/test", nil)
	req1.Header.Set(ClientIDHeader, "test-client")
	w1 := httptest.NewRecorder()

	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.JSONEq(t, `{"status":"success"}`, w1.Body.String())

	// Second request should be rate limited
	req2 := httptest.NewRequest("GET", "/api/test", nil)
	req2.Header.Set(ClientIDHeader, "test-client")
	w2 := httptest.NewRecorder()

	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.JSONEq(t, `{"error":"Rate limit exceeded","retry_after":"60 seconds"}`, w2.Body.String())
}
