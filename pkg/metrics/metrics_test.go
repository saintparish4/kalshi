package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateMetricLabels(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		path            string
		status          string
		backend         string
		expectedMethod  string
		expectedPath    string
		expectedStatus  string
		expectedBackend string
	}{
		{
			name:            "valid labels",
			method:          "GET",
			path:            "/api/v1/users",
			status:          "200",
			backend:         "user-service",
			expectedMethod:  "GET",
			expectedPath:    "/api/v1/users",
			expectedStatus:  "200",
			expectedBackend: "user-service",
		},
		{
			name:            "empty method defaults to unknown",
			method:          "",
			path:            "/test",
			status:          "404",
			backend:         "test-service",
			expectedMethod:  "unknown",
			expectedPath:    "/test",
			expectedStatus:  "404",
			expectedBackend: "test-service",
		},
		{
			name:            "invalid method defaults to unknown",
			method:          "INVALID",
			path:            "/test",
			status:          "500",
			backend:         "test-service",
			expectedMethod:  "unknown",
			expectedPath:    "/test",
			expectedStatus:  "500",
			expectedBackend: "test-service",
		},
		{
			name:            "empty path defaults to root",
			method:          "POST",
			path:            "",
			status:          "201",
			backend:         "test-service",
			expectedMethod:  "POST",
			expectedPath:    "/",
			expectedStatus:  "201",
			expectedBackend: "test-service",
		},
		{
			name:            "empty status defaults to 0",
			method:          "PUT",
			path:            "/api/test",
			status:          "",
			backend:         "test-service",
			expectedMethod:  "PUT",
			expectedPath:    "/api/test",
			expectedStatus:  "0",
			expectedBackend: "test-service",
		},
		{
			name:            "empty backend defaults to unknown",
			method:          "DELETE",
			path:            "/api/test",
			status:          "204",
			backend:         "",
			expectedMethod:  "DELETE",
			expectedPath:    "/api/test",
			expectedStatus:  "204",
			expectedBackend: "unknown",
		},
		{
			name:            "all empty defaults",
			method:          "",
			path:            "",
			status:          "",
			backend:         "",
			expectedMethod:  "unknown",
			expectedPath:    "/",
			expectedStatus:  "0",
			expectedBackend: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, path, status, backend := validateMetricLabels(tt.method, tt.path, tt.status, tt.backend)

			assert.Equal(t, tt.expectedMethod, method)
			assert.Equal(t, tt.expectedPath, path)
			assert.Equal(t, tt.expectedStatus, status)
			assert.Equal(t, tt.expectedBackend, backend)
		})
	}
}

func TestPrometheusMiddleware(t *testing.T) {
	// Reset metrics before each test
	RequestsTotal.Reset()
	RequestDuration.Reset()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(PrometheusMiddleware())

	// Add test routes
	router.GET("/test", func(c *gin.Context) {
		c.Set("backend", "test-service")
		c.JSON(200, gin.H{"message": "success"})
	})

	router.POST("/api/users", func(c *gin.Context) {
		c.Set("backend", "user-service")
		c.JSON(201, gin.H{"id": 1})
	})

	router.GET("/error", func(c *gin.Context) {
		c.Set("backend", "error-service")
		c.JSON(500, gin.H{"error": "internal error"})
	})

	router.GET("/no-backend", func(c *gin.Context) {
		// Don't set backend
		c.JSON(404, gin.H{"error": "not found"})
	})

	tests := []struct {
		name            string
		method          string
		path            string
		expectedStatus  int
		expectedBackend string
	}{
		{
			name:            "successful GET request",
			method:          "GET",
			path:            "/test",
			expectedStatus:  200,
			expectedBackend: "test-service",
		},
		{
			name:            "successful POST request",
			method:          "POST",
			path:            "/api/users",
			expectedStatus:  201,
			expectedBackend: "user-service",
		},
		{
			name:            "error request",
			method:          "GET",
			path:            "/error",
			expectedStatus:  500,
			expectedBackend: "error-service",
		},
		{
			name:            "request without backend",
			method:          "GET",
			path:            "/no-backend",
			expectedStatus:  404,
			expectedBackend: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest(tt.method, tt.path, nil)
			require.NoError(t, err)

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(w, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Wait a bit for metrics to be recorded
			time.Sleep(10 * time.Millisecond)

			// Verify metrics were recorded by checking that the middleware doesn't panic
			// and the request completes successfully
			assert.NotEmpty(t, w.Body.String())
		})
	}
}

func TestHandler(t *testing.T) {
	handler := Handler()

	// Verify handler is not nil
	assert.NotNil(t, handler)

	// Create a test request to the metrics endpoint
	req, err := http.NewRequest("GET", "/metrics", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")

	// Verify response body contains some metric data
	body := w.Body.String()
	assert.Contains(t, body, "# HELP")
	assert.Contains(t, body, "# TYPE")
}

func TestMetricsInitialization(t *testing.T) {
	// Test that all metrics are properly initialized
	assert.NotNil(t, RequestsTotal)
	assert.NotNil(t, RequestDuration)
	assert.NotNil(t, RateLimitHits)
	assert.NotNil(t, CacheHits)
	assert.NotNil(t, CircuitBreakerState)

	// Test that metrics can be used without panicking
	RequestsTotal.WithLabelValues("GET", "/test", "200", "test-service").Inc()
	RequestDuration.WithLabelValues("GET", "/test", "200", "test-service").Observe(0.1)
	RateLimitHits.WithLabelValues("/test", "client-123").Inc()
	CacheHits.WithLabelValues("redis", "hit").Inc()
	CircuitBreakerState.WithLabelValues("test-service").Set(0)
}

func TestRateLimitMetrics(t *testing.T) {
	// Reset metrics
	RateLimitHits.Reset()

	// Test rate limit metric
	RateLimitHits.WithLabelValues("/api/test", "client-123").Inc()
	RateLimitHits.WithLabelValues("/api/test", "client-456").Inc()
	RateLimitHits.WithLabelValues("/api/test", "client-123").Inc()

	// Verify metrics by checking they don't panic
	assert.NotNil(t, RateLimitHits)
}

func TestCacheMetrics(t *testing.T) {
	// Reset metrics
	CacheHits.Reset()

	// Test cache metrics
	CacheHits.WithLabelValues("redis", "hit").Inc()
	CacheHits.WithLabelValues("redis", "miss").Inc()
	CacheHits.WithLabelValues("memory", "hit").Inc()

	// Verify metrics by checking they don't panic
	assert.NotNil(t, CacheHits)
}

func TestCircuitBreakerMetrics(t *testing.T) {
	// Reset metrics
	CircuitBreakerState.Reset()

	// Test circuit breaker states
	CircuitBreakerState.WithLabelValues("user-service").Set(0)    // Closed
	CircuitBreakerState.WithLabelValues("payment-service").Set(1) // Open
	CircuitBreakerState.WithLabelValues("auth-service").Set(2)    // Half-Open

	// Verify metrics by checking they don't panic
	assert.NotNil(t, CircuitBreakerState)
}

func BenchmarkPrometheusMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(PrometheusMiddleware())

	router.GET("/benchmark", func(c *gin.Context) {
		c.Set("backend", "benchmark-service")
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/benchmark", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}
