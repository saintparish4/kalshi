package middleware

import (
	"time"

	"kalshi/pkg/logger"

	"github.com/gin-gonic/gin"
)

const (
	// Context keys for logging
	LogContextUserID     = "user_id"
	LogContextAuthMethod = "auth_method"
	LogContextBackend    = "backend"

	// Default values
	LogDefaultUserID     = "anonymous"
	LogDefaultAuthMethod = "none"
)

// RequestLogging logs all HTTP requests
func RequestLogging(log *logger.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Get additional context if available
		userID := LogDefaultUserID
		authMethod := LogDefaultAuthMethod
		backend := ""

		if param.Keys != nil {
			if uid, exists := param.Keys[LogContextUserID]; exists {
				if u, ok := uid.(string); ok {
					userID = u
				}
			}
			if auth, exists := param.Keys[LogContextAuthMethod]; exists {
				if a, ok := auth.(string); ok {
					authMethod = a
				}
			}
			if be, exists := param.Keys[LogContextBackend]; exists {
				if b, ok := be.(string); ok {
					backend = b
				}
			}
		}

		// Log the request
		fields := map[string]interface{}{
			"timestamp":   param.TimeStamp.Format(time.RFC3339),
			"status":      param.StatusCode,
			"latency":     param.Latency.String(),
			"client_ip":   param.ClientIP,
			"method":      param.Method,
			"path":        param.Path,
			"user_agent":  param.Request.UserAgent(),
			"user_id":     userID,
			"auth_method": authMethod,
		}

		if backend != "" {
			fields["backend"] = backend
		}

		if param.ErrorMessage != "" {
			fields["error"] = param.ErrorMessage
		}

		// Log based on status code
		if param.StatusCode >= 500 {
			log.WithFields(fields).Error("Request processed with server error")
		} else if param.StatusCode >= 400 {
			log.WithFields(fields).Warn("Request processed with client error")
		} else {
			log.WithFields(fields).Info("Request processed successfully")
		}

		return ""
	})
}

// StructuredLogging provides more detailed logging with custom fields
func StructuredLogging(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get client IP
		clientIP := c.ClientIP()

		// Get user info from context
		userID := LogDefaultUserID
		authMethod := LogDefaultAuthMethod
		if uid, exists := c.Get(LogContextUserID); exists {
			if u, ok := uid.(string); ok {
				userID = u
			}
		}
		if auth, exists := c.Get(LogContextAuthMethod); exists {
			if a, ok := auth.(string); ok {
				authMethod = a
			}
		}

		// Build log entry
		fields := map[string]interface{}{
			"status":      c.Writer.Status(),
			"method":      c.Request.Method,
			"path":        path,
			"query":       raw,
			"ip":          clientIP,
			"user_agent":  c.Request.UserAgent(),
			"user_id":     userID,
			"auth_method": authMethod,
			"latency":     latency.String(),
			"size":        c.Writer.Size(),
		}

		// Add error if present
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// Log based on status
		status := c.Writer.Status()
		if status >= 500 {
			log.WithFields(fields).Error("Server error")
		} else if status >= 400 {
			log.WithFields(fields).Warn("Client error")
		} else {
			log.WithFields(fields).Info("Request completed")
		}
	}
}
