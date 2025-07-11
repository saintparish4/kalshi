package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	// CORSMaxAge is the maximum age for CORS preflight requests (24 hours)
	CORSMaxAge = "86400"
	// DefaultAllowedMethods are the default HTTP methods allowed in CORS
	DefaultAllowedMethods = "GET, POST, PUT, DELETE, OPTIONS, PATCH"
	// DefaultAllowedHeaders are the default headers allowed in CORS
	DefaultAllowedHeaders = "Origin, Content-Type, Authorization, X-API-Key, X-Client-ID"
	// DefaultExposeHeaders are the default headers exposed in CORS responses
	DefaultExposeHeaders = "Content-Length, X-RateLimit-Limit, X-RateLimit-Remaining, X-Cache"
)

// CORS handles Cross-Origin Resource Sharing with permissive settings.
// For production use, consider using CORSWithConfig or RestrictedCORS
// with specific allowed origins.
func CORS() gin.HandlerFunc {
	return CORSWithConfig(CORSConfig{
		AllowAllOrigins: true,
	})
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowAllOrigins bool
	AllowedOrigins  []string
}

// CORSWithConfig handles CORS with custom configuration
func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Set CORS headers
		if config.AllowAllOrigins {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if len(config.AllowedOrigins) > 0 {
			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range config.AllowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}

			if allowed {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}

		c.Header("Access-Control-Allow-Methods", DefaultAllowedMethods)
		c.Header("Access-Control-Allow-Headers", DefaultAllowedHeaders)
		c.Header("Access-Control-Expose-Headers", DefaultExposeHeaders)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", CORSMaxAge)

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RestrictedCORS allows only specific origins
func RestrictedCORS(allowedOrigins []string) gin.HandlerFunc {
	return CORSWithConfig(CORSConfig{
		AllowAllOrigins: false,
		AllowedOrigins:  allowedOrigins,
	})
}
