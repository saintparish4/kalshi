package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ValidateContentType middleware validates that the request has an acceptable content type
func ValidateContentType(allowedTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip validation for GET, HEAD, DELETE, OPTIONS requests
		if c.Request.Method == "GET" ||
			c.Request.Method == "HEAD" ||
			c.Request.Method == "DELETE" ||
			c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if contentType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Content-Type header is required",
			})
			c.Abort()
			return
		}

		// Remove charset and other parameters
		contentType = strings.Split(contentType, ";")[0]
		contentType = strings.TrimSpace(contentType)

		// Check if content type is allowed
		for _, allowedType := range allowedTypes {
			if contentType == allowedType {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error":         "Unsupported content type",
			"allowed_types": allowedTypes,
			"received_type": contentType,
		})
		c.Abort()
	}
}

// ValidateHeaders middleware validates that required headers are present
func ValidateHeaders(requiredHeaders ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, header := range requiredHeaders {
			if c.GetHeader(header) == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":  "Required header missing",
					"header": header,
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// SecurityHeaders adds security-related headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")

		// Remove potentially sensitive headers
		c.Header("X-Powered-By", "")

		c.Next()
	}
}

// RemoveServerHeader removes the Server header from responses
func RemoveServerHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Server", "")
		c.Next()
	}
}
