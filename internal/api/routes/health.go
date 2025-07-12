package routes

import (
	"kalshi/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

// setupHealthRoutes configures health check endpoints
func setupHealthRoutes(router *gin.Engine, cfg *RouterConfig) {
	healthHandler := handlers.NewHealthHandler(cfg.Gateway)

	// Standard health endpoints
	health := router.Group("/health")
	{
		health.GET("", healthHandler.Health)
		health.GET("/", healthHandler.Health)
		health.GET("/check", healthHandler.Health)
	}

	// Kubernetes-style health endpoints
	k8s := router.Group("")
	{
		k8s.GET("/healthz", healthHandler.Health)
		k8s.GET("/ready", healthHandler.Readiness)
		k8s.GET("/readiness", healthHandler.Readiness)
		k8s.GET("/live", healthHandler.Liveness)
		k8s.GET("/liveness", healthHandler.Liveness)
	}

	// Detailed health information
	detailed := router.Group("/health")
	{
		detailed.GET("/detailed", healthHandler.DetailedHealth)
		detailed.GET("/full", healthHandler.DetailedHealth)
		detailed.GET("/status", healthHandler.DetailedHealth)
	}
}

// setupPublicRoutes configures publicly accessible routes
func setupPublicRoutes(router *gin.Engine, cfg *RouterConfig) {
	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "Kalshi API Gateway",
			"version": "1.0.0",
			"status":  "running",
			"docs":    "/docs",
			"health":  "/health",
			"metrics": "/metrics",
		})
	})

	// API documentation
	router.GET("/docs", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"documentation": "API Documentation",
			"endpoints": gin.H{
				"health":  "/health",
				"metrics": "/metrics",
				"api":     "/api/*",
				"admin":   "/admin/*",
			},
		})
	})

	// Favicon (to prevent 404s)
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204)
	})

	// Robots.txt
	router.GET("/robots.txt", func(c *gin.Context) {
		c.String(200, "User-agent: *\nDisallow: /")
	})
}
