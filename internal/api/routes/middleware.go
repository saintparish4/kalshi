package routes

import (
	"time"

	"kalshi/internal/api/middleware"
	"kalshi/internal/auth"
	"kalshi/internal/ratelimit"
	"kalshi/pkg/logger"

	"github.com/gin-gonic/gin"
)

// RouteMiddleware provides route-specific middleware configurations
type RouteMiddleware struct {
	logger *logger.Logger
}

// NewRouteMiddleware creates a new route middleware manager
func NewRouteMiddleware(log *logger.Logger) *RouteMiddleware {
	return &RouteMiddleware{logger: log}
}

// ForAPI returns middleware stack for API routes
func (rm *RouteMiddleware) ForAPI() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.SecurityHeaders(),
		middleware.ValidateContentType("application/json"),
		middleware.ValidateHeaders("Content-Type"),
	}
}

// ForUpload returns middleware stack for file upload routes
func (rm *RouteMiddleware) ForUpload() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.SecurityHeaders(),
		middleware.Timeout(5 * time.Minute),
		middleware.ValidateContentType("multipart/form-data", "application/octet-stream"),
	}
}

// ForStreaming returns middleware stack for streaming routes
func (rm *RouteMiddleware) ForStreaming() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.SecurityHeaders(),
		// No timeout for streaming
	}
}

// ForAdmin returns middleware stack for admin routes
func (rm *RouteMiddleware) ForAdmin() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.SecurityHeaders(),
		middleware.RemoveServerHeader(),
	}
}

// ForPublic returns middleware stack for public routes
func (rm *RouteMiddleware) ForPublic() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.SecurityHeaders(),
		middleware.RemoveServerHeader(),
	}
}

// WithRateLimit adds rate limiting with custom settings
func (rm *RouteMiddleware) WithRateLimit(limiter *ratelimit.Limiter, customRate int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if customRate > 0 {
			c.Set("custom_rate_limit", customRate)
		}
		middleware.RateLimit(limiter, rm.logger)(c)
	}
}

// WithAuth adds authentication with fallback options
func (rm *RouteMiddleware) WithAuth(
	jwtManager *auth.JWTManager,
	apiKeyManager *auth.APIKeyManager,
	header string,
	required bool,
) gin.HandlerFunc {
	if required {
		return middleware.JWTAuth(jwtManager, rm.logger)
	}
	return middleware.OptionalAuth(jwtManager, apiKeyManager, header, rm.logger)
}
