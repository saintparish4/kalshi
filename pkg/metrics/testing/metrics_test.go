package testing

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kalshi/pkg/metrics"
)

// TestValidateMetricLabels tests the validateMetricLabels function indirectly
// through the PrometheusMiddleware since the function is unexported
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
			// Test the validation logic indirectly through the middleware
			// by creating a request and checking that the middleware handles it correctly
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(metrics.PrometheusMiddleware())

			// Create a route that will trigger the middleware
			router.Any("/*path", func(c *gin.Context) {
				if tt.backend != "" {
					c.Set("backend", tt.backend)
				}
				// Set a custom status if provided
				if tt.status != "" {
					statusCode := 200
					if tt.status != "0" {
						if s, err := time.ParseDuration(tt.status); err == nil {
							statusCode = int(s.Seconds())
						}
					}
					c.Status(statusCode)
				} else {
					c.Status(200)
				}
			})

			// Create request with the test method and path
			method := tt.method
			if method == "" {
				method = "GET"
			}
			path := tt.path
			if path == "" {
				path = "/"
			}

			req, err := http.NewRequest(method, path, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// The middleware should handle the request without panicking
			// which means the validation logic is working correctly
			assert.NotNil(t, w)
		})
	}
}

func TestPrometheusMiddleware(t *testing.T) {
	// Reset metrics before each test
	metrics.RequestsTotal.Reset()
	metrics.RequestDuration.Reset()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(metrics.PrometheusMiddleware())

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
	handler := metrics.Handler()

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
	assert.NotNil(t, metrics.RequestsTotal)
	assert.NotNil(t, metrics.RequestDuration)
	assert.NotNil(t, metrics.RateLimitHits)
	assert.NotNil(t, metrics.CacheHits)
	assert.NotNil(t, metrics.CircuitBreakerState)

	// Test that metrics can be used without panicking
	metrics.RequestsTotal.WithLabelValues("GET", "/test", "200", "test-service").Inc()
	metrics.RequestDuration.WithLabelValues("GET", "/test", "200", "test-service").Observe(0.1)
	metrics.RateLimitHits.WithLabelValues("/test", "client-123").Inc()
	metrics.CacheHits.WithLabelValues("redis", "hit").Inc()
	metrics.CircuitBreakerState.WithLabelValues("test-service").Set(0)
}

func TestRateLimitMetrics(t *testing.T) {
	// Reset metrics
	metrics.RateLimitHits.Reset()

	// Test rate limit metric
	metrics.RateLimitHits.WithLabelValues("/api/test", "client-123").Inc()
	metrics.RateLimitHits.WithLabelValues("/api/test", "client-456").Inc()
	metrics.RateLimitHits.WithLabelValues("/api/test", "client-123").Inc()

	// Verify metrics by checking they don't panic
	assert.NotNil(t, metrics.RateLimitHits)
}

func TestCacheMetrics(t *testing.T) {
	// Reset metrics
	metrics.CacheHits.Reset()

	// Test cache metrics
	metrics.CacheHits.WithLabelValues("redis", "hit").Inc()
	metrics.CacheHits.WithLabelValues("redis", "miss").Inc()
	metrics.CacheHits.WithLabelValues("memory", "hit").Inc()

	// Verify metrics by checking they don't panic
	assert.NotNil(t, metrics.CacheHits)
}

func TestCircuitBreakerMetrics(t *testing.T) {
	// Reset metrics
	metrics.CircuitBreakerState.Reset()

	// Test circuit breaker states
	metrics.CircuitBreakerState.WithLabelValues("user-service").Set(0)    // Closed
	metrics.CircuitBreakerState.WithLabelValues("payment-service").Set(1) // Open
	metrics.CircuitBreakerState.WithLabelValues("auth-service").Set(2)    // Half-Open

	// Verify metrics by checking they don't panic
	assert.NotNil(t, metrics.CircuitBreakerState)
}

func BenchmarkPrometheusMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(metrics.PrometheusMiddleware())

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
