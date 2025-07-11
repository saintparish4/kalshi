package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"kalshi/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Recovery handles panics and recovers gracefully
func Recovery(log *logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		stack := debug.Stack()

		log.WithFields(map[string]interface{}{
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"panic":      err,
			"stack":      string(stack),
		}).Error("Panic recovered")

		// Only include request_id if it was provided in the header
		requestID := c.GetHeader("X-Request-ID")

		response := gin.H{"error": "Internal server error", "request_id": requestID}
		c.JSON(http.StatusInternalServerError, response)
	})
}

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	// Simple request ID generation
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
