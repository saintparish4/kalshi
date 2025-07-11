package middleware
import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"kalshi/internal/auth"
)

func JWTAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer"
		parts := string.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := jwtManager.VerifyToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func APIKeyAuth(apiKeyManager *auth.APIKeyManager, header string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(header)
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is required"})
			c.Abort()
			return 
		}

		keyInfo, err := apiKeyManager.ValidateAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		// Store key info in context
		c.Set("user_id", keyInfo.UserID)
		c.Set("rate_limit", keyInfo.RateLimit)
		c.Next()
	}

}