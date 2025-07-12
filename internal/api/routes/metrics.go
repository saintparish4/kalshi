package routes

import (
	"kalshi/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

// setupMetricsRoutes configures metrics and monitoring endpoints
func setupMetricsRoutes(router *gin.Engine, cfg *RouterConfig) {
	metricsHandler := handlers.NewMetricsHandler()

	// Standard Prometheus metrics
	router.GET("/metrics", handlers.MetricsHandler())

	// Custom metrics endpoints
	metrics := router.Group("/metrics")
	{
		metrics.GET("/info", metricsHandler.GetMetrics)
		// Note: Other custom metrics endpoints would need to be implemented
		// in the CustomMetricsHandler struct
	}

	// Metrics in different formats
	export := router.Group("/export")
	{
		export.GET("/prometheus", handlers.MetricsHandler())
		// Note: JSON and CSV endpoints would need to be implemented
	}

	// Real-time metrics (could be WebSocket)
	// Note: Real-time endpoints would need to be implemented in CustomMetricsHandler
}
