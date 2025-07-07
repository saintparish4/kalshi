// Package metrics provides Prometheus metrics collection for the Kalshi API gateway.
// It includes request metrics, rate limiting, caching, and circuit breaker monitoring.
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kalshi_requests_total",
			Help: "Total number of requests processed",
		},
		[]string{"method", "path", "status", "backend"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kalshi_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status", "backend"},
	)

	RateLimitHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kalshi_rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"path", "client_id"},
	)

	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kalshi_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type", "hit_type"},
	)

	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kalshi_circuit_breaker_state",
			Help: "Circuit breaker state (0: Closed, 1: Open, 2: Half-Open)",
		},
		[]string{"backend"},
	)
)

// validateMetricLabels ensures all label values are valid and non-empty.
// Returns sanitized values that are safe for Prometheus metrics.
func validateMetricLabels(method, path, status, backend string) (string, string, string, string) {
	// Ensure method is valid HTTP method or default to "unknown"
	if method == "" || (method != "GET" && method != "POST" && method != "PUT" &&
		method != "DELETE" && method != "PATCH" && method != "HEAD" && method != "OPTIONS") {
		method = "unknown"
	}

	// Ensure path is not empty
	if path == "" {
		path = "/"
	}

	// Ensure status is a valid HTTP status code or default to "0"
	if status == "" {
		status = "0"
	}

	// Ensure backend is not empty
	if backend == "" {
		backend = "unknown"
	}

	return method, path, status, backend
}

// PrometheusMiddleware returns a Gin middleware that collects HTTP request metrics.
// It tracks request counts, durations, and status codes for monitoring API performance.
func PrometheusMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())

		// Ensure backend label has a value to prevent inconsistent metrics
		backend := c.GetString("backend")
		if backend == "" {
			backend = "unknown"
		}

		// Validate and sanitize all metric labels
		method, path, statusCode, backendName := validateMetricLabels(
			c.Request.Method,
			c.FullPath(),
			status,
			backend,
		)

		RequestsTotal.WithLabelValues(
			method,
			path,
			statusCode,
			backendName,
		).Inc()

		RequestDuration.WithLabelValues(
			method,
			path,
			statusCode,
			backendName,
		).Observe(duration.Seconds())
	})
}

// Handler returns an HTTP handler that serves Prometheus metrics at the /metrics endpoint.
// This handler can be mounted on your server to expose metrics for scraping.
func Handler() http.Handler {
	return promhttp.Handler()
}
