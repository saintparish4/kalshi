package handlers

import (
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

// Custom metrics endpoints for debugging
type CustomMetricsHandler struct{}

func NewCustomMetricsHandler() *CustomMetricsHandler {
	return &CustomMetricsHandler{}
}

// GetMetrics returns custom metrics in JSON format
func (m *CustomMetricsHandler) GetMetrics(c *gin.Context) {
	// This could return custom metrics for debugging
	c.JSON(200, gin.H{
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
