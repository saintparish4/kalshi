package middleware

import (
	"net/http"
	"strconv"
	"time"

	"kalshi/internal/ratelimit"
	"kalshi/pkg/logger"

	"github.com/gin-gonic/gin"
)

const (
	// Rate limit constants
	DefaultRateLimit         = 60 // requests per minute
	RateLimitHeader          = "X-RateLimit-Limit"
	RateLimitRemainingHeader = "X-RateLimit-Remaining"
	RateLimitResetHeader     = "X-RateLimit-Reset"
	ClientIDHeader           = "X-Client-ID"
)

// RateLimit enforces rate limiting based on client identifier and path.
// It supports custom rate limits set by authentication middleware
// and provides rate limit headers in responses.
func RateLimit(limiter *ratelimit.Limiter, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier
		clientID := getClientID(c)
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Get custom rate limit from context (set by auth middleware)
		customRate := 0
		if rate, exists := c.Get("rate_limit"); exists {
			if r, ok := rate.(int); ok {
				customRate = r
			}
		}

		// Check rate limit
		allowed, err := limiter.Allow(c.Request.Context(), clientID, path)
		if err != nil {
			log.Error("Rate limit check failed",
				"error", err,
				"client_id", clientID,
				"path", path,
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Rate limit check failed",
			})
			c.Abort()
			return
		}

		if !allowed {
			// Add rate limit headers
			effectiveRate := customRate
			if effectiveRate == 0 {
				effectiveRate = DefaultRateLimit // Default rate limit
			}

			c.Header(RateLimitHeader, strconv.Itoa(effectiveRate))
			c.Header(RateLimitRemainingHeader, "0")
			c.Header(RateLimitResetHeader, strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))

			log.Warn("Rate limit exceeded",
				"client_id", clientID,
				"path", path,
				"rate_limit", effectiveRate,
			)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": "60 seconds",
			})
			c.Abort()
			return
		}

		// Add rate limit headers for successful requests
		effectiveRate := customRate
		if effectiveRate == 0 {
			effectiveRate = DefaultRateLimit // Default rate limit
		}
		c.Header(RateLimitHeader, strconv.Itoa(effectiveRate))

		c.Next()
	}
}

// getClientID determines the client identifier for rate limiting
func getClientID(c *gin.Context) string {
	// Try to get user ID from context (set by auth middleware)
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok && uid != "anonymous" {
			return uid
		}
	}

	// Try to get from custom header
	if clientID := c.GetHeader(ClientIDHeader); clientID != "" {
		return clientID
	}

	// Fall back to IP address
	return c.ClientIP()
}
