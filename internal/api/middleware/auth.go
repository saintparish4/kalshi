package middleware

import (
	"net/http"
	"strings"

	"kalshi/internal/auth"
	"kalshi/pkg/logger"

	"github.com/gin-gonic/gin"
)

const (
	// Auth error messages
	ErrMissingAuthHeader = "Authorization header required"
	ErrInvalidAuthFormat = "Invalid authorization header format. Use: Bearer <token>"
	ErrInvalidToken      = "Invalid or expired token"
	ErrMissingAPIKey     = "API key required"
	ErrInvalidAPIKey     = "Invalid API key"

	// Context keys
	ContextUserID     = "user_id"
	ContextRole       = "role"
	ContextAuthMethod = "auth_method"
	ContextRateLimit  = "rate_limit"

	// Auth methods
	AuthMethodJWT    = "jwt"
	AuthMethodAPIKey = "api_key"
	AuthMethodNone   = "none"

	// Default values
	AnonymousUser = "anonymous"
)

// JWTAuth validates JWT tokens from the Authorization header.
// It expects tokens in the format "Bearer <token>" and stores user information
// in the request context for downstream handlers.
func JWTAuth(jwtManager *auth.JWTManager, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("Missing authorization header", "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": ErrMissingAuthHeader,
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Warn("Invalid authorization header format", "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": ErrInvalidAuthFormat,
			})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := jwtManager.ValidateToken(c.Request.Context(), token)
		if err != nil {
			log.Warn("Invalid JWT token", "error", err, "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": ErrInvalidToken,
			})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextRole, claims.Role)
		c.Set(ContextAuthMethod, AuthMethodJWT)

		log.Debug("JWT authentication successful",
			"user_id", claims.UserID,
			"role", claims.Role,
		)

		c.Next()
	}
}

// APIKeyAuth validates API keys from the specified header or query parameter.
// It stores user information and rate limit settings in the request context
// for downstream handlers.
func APIKeyAuth(apiKeyManager *auth.APIKeyManager, header string, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(header)
		if apiKey == "" {
			// Also check query parameter as fallback
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			log.Warn("Missing API key", "ip", c.ClientIP(), "header", header)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": ErrMissingAPIKey + " in header: " + header,
			})
			c.Abort()
			return
		}

		keyInfo, err := apiKeyManager.ValidateAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			log.Warn("Invalid API key", "error", err, "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": ErrInvalidAPIKey,
			})
			c.Abort()
			return
		}

		// Store key info in context
		c.Set(ContextUserID, keyInfo.UserID)
		c.Set(ContextRateLimit, keyInfo.RateLimit)
		c.Set(ContextAuthMethod, AuthMethodAPIKey)

		log.Debug("API key authentication successful",
			"user_id", keyInfo.UserID,
			"rate_limit", keyInfo.RateLimit,
		)

		c.Next()
	}
}

// OptionalAuth tries JWT authentication first, then API key authentication.
// If neither succeeds, it allows the request to continue as anonymous.
// This is useful for endpoints that support both authenticated and unauthenticated access.
func OptionalAuth(jwtManager *auth.JWTManager, apiKeyManager *auth.APIKeyManager, apiKeyHeader string, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try JWT first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token := parts[1]
				if claims, err := jwtManager.ValidateToken(c.Request.Context(), token); err == nil {
					c.Set(ContextUserID, claims.UserID)
					c.Set(ContextRole, claims.Role)
					c.Set(ContextAuthMethod, AuthMethodJWT)
					c.Next()
					return
				}
			}
		}

		// Try API key
		apiKey := c.GetHeader(apiKeyHeader)
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey != "" {
			if keyInfo, err := apiKeyManager.ValidateAPIKey(c.Request.Context(), apiKey); err == nil {
				c.Set(ContextUserID, keyInfo.UserID)
				c.Set(ContextRateLimit, keyInfo.RateLimit)
				c.Set(ContextAuthMethod, AuthMethodAPIKey)
				c.Next()
				return
			}
		}

		// No authentication - continue as anonymous
		c.Set(ContextUserID, AnonymousUser)
		c.Set(ContextAuthMethod, AuthMethodNone)
		c.Next()
	}
}
