package routes

import (
	"kalshi/internal/api/handlers"
	"kalshi/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// setupAdminRoutes configures administrative endpoints
func setupAdminRoutes(router *gin.Engine, cfg *RouterConfig) {
	adminHandler := handlers.NewAdminHandler(cfg.Gateway, cfg.Logger)
	proxyHandler := handlers.NewProxyHandler(cfg.Gateway, cfg.Config, cfg.Logger)

	admin := router.Group("/admin")

	// Optional authentication for admin routes
	admin.Use(middleware.OptionalAuth(
		cfg.JWTManager,
		cfg.APIKeyManager,
		cfg.Config.Auth.APIKey.Header,
		cfg.Logger,
	))

	// Backend management
	backends := admin.Group("/backends")
	{
		backends.GET("", adminHandler.GetBackends)
		backends.GET("/:name", adminHandler.GetBackend)
		backends.POST("/:name/health", adminHandler.CheckBackendHealth)
		backends.PUT("/:name/enable", adminHandler.EnableBackend)
		backends.PUT("/:name/disable", adminHandler.DisableBackend)
	}

	// Circuit breaker management
	circuits := admin.Group("/circuits")
	{
		circuits.GET("", adminHandler.GetCircuitBreakers)
		circuits.GET("/:backend", adminHandler.GetCircuitBreaker)
		circuits.POST("/:backend/reset", adminHandler.ResetCircuitBreaker)
		circuits.PUT("/:backend/open", adminHandler.OpenCircuitBreaker)
		circuits.PUT("/:backend/close", adminHandler.CloseCircuitBreaker)
	}

	// Route management
	routes := admin.Group("/routes")
	{
		routes.GET("", proxyHandler.GetRoutes)
		routes.GET("/:id", proxyHandler.GetRoute)
		routes.POST("/reload", proxyHandler.ReloadRoutes)
	}

	// Cache management
	cache := admin.Group("/cache")
	{
		cache.GET("/stats", adminHandler.GetCacheStats)
		cache.DELETE("/clear", adminHandler.ClearCache)
		cache.DELETE("/:key", adminHandler.DeleteCacheKey)
		cache.GET("/:key", adminHandler.GetCacheKey)
	}

	// Rate limiting management
	rateLimit := admin.Group("/ratelimit")
	{
		rateLimit.GET("/stats", adminHandler.GetRateLimitStats)
		rateLimit.DELETE("/reset/:clientId", adminHandler.ResetRateLimit)
		rateLimit.GET("/status/:clientId", adminHandler.GetRateLimitStatus)
	}

	// System information
	system := admin.Group("/system")
	{
		system.GET("/stats", adminHandler.GetStats)
		system.GET("/info", adminHandler.GetSystemInfo)
		system.GET("/config", adminHandler.GetConfig)
		system.POST("/reload", adminHandler.ReloadConfig)
	}

	// Logs and debugging
	debug := admin.Group("/debug")
	{
		debug.GET("/goroutines", adminHandler.GetGoroutines)
		debug.GET("/memory", adminHandler.GetMemoryStats)
		debug.GET("/gc", adminHandler.TriggerGC)
		debug.GET("/pprof/*path", adminHandler.HandlePprof)
	}
}
