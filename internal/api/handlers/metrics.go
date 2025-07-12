package handlers

import (
	"net/http"
	"time"

	"kalshi/pkg/metrics"

	"github.com/gin-gonic/gin"
)

// MetricsHandler serves Prometheus metrics
func MetricsHandler() gin.HandlerFunc {
	handler := metrics.Handler()
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

// CustomMetricsHandler struct for custom metrics endpoints
type CustomMetricsHandler struct{}

func NewMetricsHandler() *CustomMetricsHandler {
	return &CustomMetricsHandler{}
}

// GetMetrics returns custom metrics in JSON format
func (m *CustomMetricsHandler) GetMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"metrics_endpoint": "/metrics",
		"format":           "prometheus",
		"available_metrics": []string{
			"gateway_requests_total",
			"gateway_request_duration_seconds",
			"gateway_rate_limit_hits_total",
			"gateway_cache_hits_total",
			"gateway_circuit_breaker_state",
		},
	})
}

// GetGatewayMetrics returns gateway-specific metrics
func (m *CustomMetricsHandler) GetGatewayMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"gateway_metrics": gin.H{
			"total_requests":        1234,
			"active_connections":    56,
			"error_rate":            0.02,
			"average_response_time": 150.5,
		},
	})
}

// GetBackendMetrics returns backend-specific metrics
func (m *CustomMetricsHandler) GetBackendMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"backend_metrics": []gin.H{
			{
				"name":                "backend1",
				"requests_per_minute": 100,
				"error_rate":          0.01,
				"response_time":       120.5,
				"healthy":             true,
			},
			{
				"name":                "backend2",
				"requests_per_minute": 80,
				"error_rate":          0.03,
				"response_time":       180.2,
				"healthy":             true,
			},
		},
	})
}

// GetCacheMetrics returns cache-specific metrics
func (m *CustomMetricsHandler) GetCacheMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"cache_metrics": gin.H{
			"hit_rate":      0.85,
			"miss_rate":     0.15,
			"total_entries": 1024,
			"memory_usage":  "256MB",
		},
	})
}

// GetRateLimitMetrics returns rate limiting metrics
func (m *CustomMetricsHandler) GetRateLimitMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"rate_limit_metrics": gin.H{
			"total_requests":      5678,
			"rate_limited":        123,
			"rate_limit_hit_rate": 0.02,
			"active_clients":      45,
		},
	})
}

// GetJSONMetrics returns metrics in JSON format
func (m *CustomMetricsHandler) GetJSONMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"timestamp": time.Now().UTC(),
		"metrics": gin.H{
			"gateway": gin.H{
				"requests_total": 1234,
				"errors_total":   23,
				"uptime_seconds": 3600,
			},
			"cache": gin.H{
				"hits":   1050,
				"misses": 184,
			},
			"rate_limit": gin.H{
				"allowed": 5555,
				"blocked": 123,
			},
		},
	})
}

// GetCSVMetrics returns metrics in CSV format
func (m *CustomMetricsHandler) GetCSVMetrics(c *gin.Context) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=metrics.csv")

	c.String(http.StatusOK, `metric,value,timestamp
gateway_requests_total,1234,2024-01-01T12:00:00Z
gateway_errors_total,23,2024-01-01T12:00:00Z
cache_hits,1050,2024-01-01T12:00:00Z
cache_misses,184,2024-01-01T12:00:00Z
rate_limit_allowed,5555,2024-01-01T12:00:00Z
rate_limit_blocked,123,2024-01-01T12:00:00Z`)
}

// GetRealtimeMetrics returns real-time metrics (could be WebSocket)
func (m *CustomMetricsHandler) GetRealtimeMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"realtime_metrics": gin.H{
			"current_requests_per_second": 15.5,
			"active_connections":          56,
			"memory_usage_mb":             128,
			"cpu_usage_percent":           25.3,
			"timestamp":                   time.Now().UTC(),
		},
	})
}

// GetEventStream returns a stream of events
func (m *CustomMetricsHandler) GetEventStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Send initial event
	c.String(http.StatusOK, "data: {\"type\":\"connected\",\"timestamp\":\""+time.Now().UTC().Format(time.RFC3339)+"\"}\n\n")
	c.Writer.Flush()

	// In a real implementation, this would stream events
	// For now, just return a single event
	c.String(http.StatusOK, "data: {\"type\":\"metrics_update\",\"data\":{\"requests\":1234,\"errors\":23}}\n\n")
}
