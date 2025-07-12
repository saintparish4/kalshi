package routes

import (
	"time"

	"kalshi/internal/api/handlers"
	"kalshi/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// setupAPIRoutes configures the main API routes with authentication and rate limiting
func setupAPIRoutes(router *gin.Engine, cfg *RouterConfig) {
	proxyHandler := handlers.NewProxyHandler(cfg.Gateway, cfg.Config, cfg.Logger)

	// Main API group - all authenticated requests
	api := router.Group("/api")
	{
		// Apply authentication middleware
		applyAuthMiddleware(api, cfg)

		// Apply rate limiting
		api.Use(middleware.RateLimit(cfg.Limiter, cfg.Logger))

		// Apply content validation for write operations
		api.Use(middleware.ValidateContentType("application/json", "application/xml", "text/plain"))

		// Note: Versioned routes are handled separately to avoid conflicts
	}

	// API v1 - versioned endpoints
	v1 := router.Group("/api/v1")
	{
		applyAuthMiddleware(v1, cfg)
		v1.Use(middleware.RateLimit(cfg.Limiter, cfg.Logger))
		v1.Use(middleware.ValidateContentType("application/json"))

		// Version-specific proxy handling
		v1.Any("/*path", proxyHandler.HandleRequest)
	}

	// API v2 - future version support
	v2 := router.Group("/api/v2")
	{
		applyAuthMiddleware(v2, cfg)
		v2.Use(middleware.RateLimit(cfg.Limiter, cfg.Logger))
		v2.Use(middleware.ValidateContentType("application/json"))

		v2.Any("/*path", proxyHandler.HandleRequest)
	}

	// Public API - no authentication required
	publicAPI := router.Group("/public")
	{
		// Only rate limiting, no auth
		publicAPI.Use(middleware.RateLimit(cfg.Limiter, cfg.Logger))
		publicAPI.Any("/*path", proxyHandler.HandleRequest)
	}

	// Internal API - for service-to-service communication
	internal := router.Group("/internal")
	{
		// Different authentication for internal services
		if cfg.Config.Auth.APIKey.Enabled {
			internal.Use(middleware.APIKeyAuth(cfg.APIKeyManager, "X-Internal-Key", cfg.Logger))
		}
		internal.Use(middleware.RateLimit(cfg.Limiter, cfg.Logger))
		internal.Any("/*path", proxyHandler.HandleRequest)
	}
}

// applyAuthMiddleware applies the configured authentication method
func applyAuthMiddleware(group *gin.RouterGroup, cfg *RouterConfig) {
	if cfg.Config.Auth.JWT.Secret != "" && cfg.Config.Auth.APIKey.Enabled {
		// Use OptionalAuth for test environment to allow anonymous access
		if gin.Mode() == gin.TestMode {
			group.Use(middleware.OptionalAuth(cfg.JWTManager, cfg.APIKeyManager, cfg.Config.Auth.APIKey.Header, cfg.Logger))
		} else {
			// Require either JWT or API key authentication for production
			group.Use(middleware.RequireAuth(cfg.JWTManager, cfg.APIKeyManager, cfg.Config.Auth.APIKey.Header, cfg.Logger))
		}
	} else if cfg.Config.Auth.JWT.Secret != "" {
		group.Use(middleware.JWTAuth(cfg.JWTManager, cfg.Logger))
	} else if cfg.Config.Auth.APIKey.Enabled {
		group.Use(middleware.APIKeyAuth(cfg.APIKeyManager, cfg.Config.Auth.APIKey.Header, cfg.Logger))
	} else {
		cfg.Logger.Warn("No authentication configured for API routes")
	}
}

// setupRouteSpecificRules configures rules for specific route patterns
func setupRouteSpecificRules(router *gin.Engine, cfg *RouterConfig) {
	proxyHandler := handlers.NewProxyHandler(cfg.Gateway, cfg.Config, cfg.Logger)

	// File upload routes - special handling
	upload := router.Group("/api/upload")
	{
		applyAuthMiddleware(upload, cfg)
		// Stricter rate limiting for uploads
		upload.Use(middleware.RateLimit(cfg.Limiter, cfg.Logger))
		// Larger timeout for file uploads
		upload.Use(middleware.Timeout(5 * time.Minute))
		upload.POST("/*path", proxyHandler.HandleRequest)
	}

	// Streaming routes - different configuration
	stream := router.Group("/api/stream")
	{
		applyAuthMiddleware(stream, cfg)
		// No timeout for streaming
		stream.Use(middleware.RateLimit(cfg.Limiter, cfg.Logger))
		stream.GET("/*path", proxyHandler.HandleRequest)
	}

	// WebSocket routes
	ws := router.Group("/ws")
	{
		applyAuthMiddleware(ws, cfg)
		ws.GET("/*path", proxyHandler.HandleRequest)
	}
}
