package routes

import (
	"time"

	"github.com/gin-gonic/gin"

	"kalshi/internal/api/middleware"
	"kalshi/internal/auth"
	"kalshi/internal/config"
	"kalshi/internal/gateway"
	"kalshi/internal/ratelimit"
	"kalshi/pkg/logger"
	"kalshi/pkg/metrics"
)

// RouterConfig holds all dependencies needed for route setup
type RouterConfig struct {
	Config        *config.Config
	Gateway       *gateway.Gateway
	Limiter       *ratelimit.Limiter
	JWTManager    *auth.JWTManager
	APIKeyManager *auth.APIKeyManager
	Logger        *logger.Logger
}

// SetupRouter configures and returns the main HTTP router
func SetupRouter(cfg *RouterConfig) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.Config.Logging.Level == "production" || cfg.Config.Logging.Format == "json" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Apply global middleware
	setupGlobalMiddleware(router, cfg)

	// Setup route groups
	setupPublicRoutes(router, cfg)
	setupHealthRoutes(router, cfg)
	setupMetricsRoutes(router, cfg)
	setupAdminRoutes(router, cfg)
	setupAPIRoutes(router, cfg)
	setupAuthRoutes(router, cfg)

	return router
}

// setupGlobalMiddleware applies middleware to all routes
func setupGlobalMiddleware(router *gin.Engine, cfg *RouterConfig) {
	router.Use(middleware.Recovery(cfg.Logger))
	router.Use(middleware.RequestID())
	router.Use(middleware.RequestLogging(cfg.Logger))
	router.Use(middleware.CORS())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.Timeout(30 * time.Second))
	router.Use(metrics.PrometheusMiddleware())
}

// SetupTestRouter creates a minimal router for testing
func SetupTestRouter(cfg *RouterConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Minimal middleware for testing
	router.Use(middleware.Recovery(cfg.Logger))
	router.Use(middleware.CORS())
	router.Use(metrics.PrometheusMiddleware())

	// Setup only essential routes for testing
	setupHealthRoutes(router, cfg)
	setupAPIRoutes(router, cfg)
	setupMetricsRoutes(router, cfg) // Added for metrics endpoint in tests

	return router
}

// SetupMetricsOnlyRouter creates a router with only metrics and health
func SetupMetricsOnlyRouter(cfg *RouterConfig) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(middleware.Recovery(cfg.Logger))

	setupHealthRoutes(router, cfg)
	setupMetricsRoutes(router, cfg)

	return router
}
