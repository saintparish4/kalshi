package testing

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"kalshi/internal/circuit"
	"kalshi/internal/config"
	"kalshi/internal/gateway"
	"kalshi/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProxy(t *testing.T) {
	backendManager := gateway.NewBackendManager()
	cacheManager := NewMockCache()
	circuitManager := circuit.NewManager()
	logger := &logger.Logger{}
	config := &config.Config{}

	proxy := gateway.NewProxy(backendManager, cacheManager, circuitManager, logger, config)

	assert.NotNil(t, proxy)
}

func TestProxy_generateCacheKey(t *testing.T) {
	proxy := gateway.NewProxy(gateway.NewBackendManager(), NewMockCache(), circuit.NewManager(), &logger.Logger{}, &config.Config{})

	tests := []struct {
		name     string
		method   string
		url      string
		expected string
	}{
		{
			name:     "GET request",
			method:   "GET",
			url:      "/api/users",
			expected: "GET:/api/users",
		},
		{
			name:     "POST request",
			method:   "POST",
			url:      "/api/users",
			expected: "POST:/api/users",
		},
		{
			name:     "request with query params",
			method:   "GET",
			url:      "/api/users?id=123&name=test",
			expected: "GET:/api/users?id=123&name=test",
		},
		{
			name:     "request with path params",
			method:   "GET",
			url:      "/api/users/123",
			expected: "GET:/api/users/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, nil)
			require.NoError(t, err)

			cacheKey := proxy.GenerateCacheKey(req)
			assert.Equal(t, tt.expected, cacheKey)
		})
	}
}

func TestProxy_getCachedResponse(t *testing.T) {
	mockCache := NewMockCache()
	proxy := gateway.NewProxy(gateway.NewBackendManager(), mockCache, circuit.NewManager(), &logger.Logger{}, &config.Config{})

	// Test cache hit
	cacheKey := "GET:/api/users"
	mockCache.data[cacheKey] = []byte("cached response")

	req, err := http.NewRequest("GET", "/api/users", nil)
	require.NoError(t, err)

	cached, err := proxy.GetCachedResponse(req)
	assert.NoError(t, err)
	assert.NotNil(t, cached)
	assert.Equal(t, 200, cached.StatusCode)
	assert.Equal(t, []byte("cached response"), cached.Body)

	// Test cache miss
	req, err = http.NewRequest("GET", "/api/nonexistent", nil)
	require.NoError(t, err)

	cached, err = proxy.GetCachedResponse(req)
	assert.Error(t, err)
	assert.Nil(t, cached)
}

func TestProxy_writeCachedResponse(t *testing.T) {
	proxy := gateway.NewProxy(gateway.NewBackendManager(), NewMockCache(), circuit.NewManager(), &logger.Logger{}, &config.Config{})

	// Create cached response
	cached := &gateway.CachedResponse{
		StatusCode: 200,
		Headers: http.Header{
			"Content-Type":  []string{"application/json"},
			"Cache-Control": []string{"max-age=3600"},
		},
		Body: []byte(`{"message": "cached response"}`),
	}

	// Create response writer
	w := httptest.NewRecorder()

	// Write cached response
	proxy.WriteCachedResponse(w, cached)

	// Verify response
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "max-age=3600", w.Header().Get("Cache-Control"))
	assert.Equal(t, `{"message": "cached response"}`, w.Body.String())
}

func TestProxy_cacheResponse(t *testing.T) {
	mockCache := NewMockCache()
	proxy := gateway.NewProxy(gateway.NewBackendManager(), mockCache, circuit.NewManager(), &logger.Logger{}, &config.Config{})

	req, err := http.NewRequest("GET", "/api/users", nil)
	require.NoError(t, err)

	resp := &http.Response{
		StatusCode: 200,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(`{"message": "test"}`)),
	}

	body := []byte(`{"message": "test"}`)
	ttl := 5 * time.Minute

	proxy.CacheResponse(req, resp, body, ttl)

	// Verify cache was set
	cacheKey := proxy.GenerateCacheKey(req)
	cachedData, exists := mockCache.data[cacheKey]
	assert.True(t, exists)
	assert.Equal(t, body, cachedData)
	assert.Equal(t, ttl, mockCache.ttl[cacheKey])
}

func TestProxy_forwardRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify forwarding headers
		assert.Equal(t, "192.168.1.1:12345", r.Header.Get("X-Forwarded-For"))
		assert.Equal(t, "http", r.Header.Get("X-Forwarded-Proto"))
		assert.Equal(t, "example.com", r.Header.Get("X-Forwarded-Host"))

		// Verify original headers are preserved
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-value", r.Header.Get("X-Test-Header"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	// Create backend
	backendURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	backend := &gateway.Backend{
		Name: "test-backend",
		URL:  backendURL,
	}

	proxy := gateway.NewProxy(gateway.NewBackendManager(), NewMockCache(), circuit.NewManager(), &logger.Logger{}, &config.Config{})

	// Create request
	req, err := http.NewRequest("POST", "http://example.com/api/test", strings.NewReader(`{"data": "test"}`))
	require.NoError(t, err)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Host = "example.com"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Header", "test-value")

	// Forward request
	resp, err := proxy.ForwardRequest(req, backend)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	// Read response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, `{"message": "success"}`, string(body))
}

func TestProxy_forwardRequest_InvalidURL(t *testing.T) {
	proxy := gateway.NewProxy(gateway.NewBackendManager(), NewMockCache(), circuit.NewManager(), &logger.Logger{}, &config.Config{})

	// Create backend with invalid URL
	backend := &gateway.Backend{
		Name: "invalid-backend",
		URL:  &url.URL{Scheme: "invalid", Host: "invalid-host"},
	}

	req, err := http.NewRequest("GET", "/api/test", nil)
	require.NoError(t, err)

	// Forward request should fail
	resp, err := proxy.ForwardRequest(req, backend)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestProxy_ServeHTTP_CacheHit(t *testing.T) {
	mockCache := NewMockCache()
	backendManager := gateway.NewBackendManager()
	circuitManager := circuit.NewManager()
	logger := &logger.Logger{}
	config := &config.Config{}

	proxy := gateway.NewProxy(backendManager, mockCache, circuitManager, logger, config)

	// Add backend
	err := backendManager.AddBackend("test-backend", "http://localhost:8080", "/health", 1)
	require.NoError(t, err)

	// Set up cache
	cacheKey := "GET:/api/users"
	mockCache.data[cacheKey] = []byte("cached response")

	// Create request
	req, err := http.NewRequest("GET", "/api/users", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	// Serve HTTP request
	proxy.ServeHTTP(w, req, "test-backend", 5*time.Minute)

	// Verify cached response was served
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "cached response", w.Body.String())
}

func TestProxy_ServeHTTP_BackendNotFound(t *testing.T) {
	proxy := gateway.NewProxy(gateway.NewBackendManager(), NewMockCache(), circuit.NewManager(), &logger.Logger{}, &config.Config{})

	req, err := http.NewRequest("GET", "/api/users", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	// Try to serve request with non-existent backend
	proxy.ServeHTTP(w, req, "non-existent-backend", 5*time.Minute)

	// Should return 502 Bad Gateway
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestProxy_ServeHTTP_SuccessfulForward(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	backendManager := gateway.NewBackendManager()
	circuitManager := circuit.NewManager()
	logger := &logger.Logger{}
	config := &config.Config{}

	proxy := gateway.NewProxy(backendManager, NewMockCache(), circuitManager, logger, config)

	// Add backend with test server URL
	err := backendManager.AddBackend("test-backend", server.URL, "/health", 1)
	require.NoError(t, err)

	req, err := http.NewRequest("GET", "/api/test", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	// Serve HTTP request
	proxy.ServeHTTP(w, req, "test-backend", 5*time.Minute)

	// Verify successful response
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, `{"message": "success"}`, w.Body.String())
}

func TestProxy_ServeHTTP_CircuitBreakerOpen(t *testing.T) {
	backendManager := gateway.NewBackendManager()
	circuitManager := circuit.NewManager()
	logger := &logger.Logger{}
	config := &config.Config{}

	proxy := gateway.NewProxy(backendManager, NewMockCache(), circuitManager, logger, config)

	// Add backend
	err := backendManager.AddBackend("test-backend", "http://localhost:8080", "/health", 1)
	require.NoError(t, err)

	// Manually set backend as unhealthy to trigger circuit breaker
	err = backendManager.SetBackendHealth("test-backend", false)
	require.NoError(t, err)

	req, err := http.NewRequest("GET", "/api/test", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	// Serve HTTP request
	proxy.ServeHTTP(w, req, "test-backend", 5*time.Minute)

	// Should return 502 Bad Gateway due to unhealthy backend
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestProxy_ServeHTTP_NonGETRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	backendManager := gateway.NewBackendManager()
	circuitManager := circuit.NewManager()
	logger := &logger.Logger{}
	config := &config.Config{}

	proxy := gateway.NewProxy(backendManager, NewMockCache(), circuitManager, logger, config)

	// Add backend with test server URL
	err := backendManager.AddBackend("test-backend", server.URL, "/health", 1)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/test", strings.NewReader(`{"data": "test"}`))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	// Serve HTTP request (POST should not use cache)
	proxy.ServeHTTP(w, req, "test-backend", 5*time.Minute)

	// Verify successful response
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, `{"message": "success"}`, w.Body.String())
}

func TestProxy_ServeHTTP_NoCacheTTL(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	backendManager := gateway.NewBackendManager()
	circuitManager := circuit.NewManager()
	logger := &logger.Logger{}
	config := &config.Config{}

	proxy := gateway.NewProxy(backendManager, NewMockCache(), circuitManager, logger, config)

	// Add backend with test server URL
	err := backendManager.AddBackend("test-backend", server.URL, "/health", 1)
	require.NoError(t, err)

	req, err := http.NewRequest("GET", "/api/test", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	// Serve HTTP request with 0 TTL (no caching)
	proxy.ServeHTTP(w, req, "test-backend", 0)

	// Verify successful response
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, `{"message": "success"}`, w.Body.String())
}

func TestProxy_ServeHTTP_Non200Response(t *testing.T) {
	// Create test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	backendManager := gateway.NewBackendManager()
	circuitManager := circuit.NewManager()
	logger := &logger.Logger{}
	config := &config.Config{}

	proxy := gateway.NewProxy(backendManager, NewMockCache(), circuitManager, logger, config)

	// Add backend with test server URL
	err := backendManager.AddBackend("test-backend", server.URL, "/health", 1)
	require.NoError(t, err)

	req, err := http.NewRequest("GET", "/api/test", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	// Serve HTTP request
	proxy.ServeHTTP(w, req, "test-backend", 5*time.Minute)

	// Should return 404 from backend
	assert.Equal(t, 404, w.Code)
	assert.Equal(t, `{"error": "not found"}`, w.Body.String())
}

func TestProxy_ConcurrentRequests(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	backendManager := gateway.NewBackendManager()
	circuitManager := circuit.NewManager()
	logger := &logger.Logger{}
	config := &config.Config{}

	proxy := gateway.NewProxy(backendManager, NewMockCache(), circuitManager, logger, config)

	// Add backend with test server URL
	err := backendManager.AddBackend("test-backend", server.URL, "/health", 1)
	require.NoError(t, err)

	// Test concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			req, err := http.NewRequest("GET", "/api/test", nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			proxy.ServeHTTP(w, req, "test-backend", 5*time.Minute)

			assert.Equal(t, 200, w.Code)
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestProxy_Phase1Optimizations(t *testing.T) {
	backendManager := gateway.NewBackendManager()
	circuitManager := circuit.NewManager()
	logger := &logger.Logger{}
	config := &config.Config{}

	proxy := gateway.NewProxy(backendManager, NewMockCache(), circuitManager, logger, config)

	// Test that optimized client is created
	client1 := proxy.GetOptimizedClient()
	assert.NotNil(t, client1)

	// Test that the same client is returned (singleton pattern)
	client2 := proxy.GetOptimizedClient()
	assert.Equal(t, client1, client2)
}
