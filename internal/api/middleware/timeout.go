package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// Timeout constants
	MinTimeout = 100 * time.Millisecond
	MaxTimeout = 5 * time.Minute
)

// Timeout adds a configurable timeout to requests.
// The timeout is validated to be between MinTimeout and MaxTimeout.
// Requests that exceed the timeout return a 408 status code.
func Timeout(timeout time.Duration) gin.HandlerFunc {
	// Validate timeout duration
	if timeout < MinTimeout {
		timeout = MinTimeout
	} else if timeout > MaxTimeout {
		timeout = MaxTimeout
	}

	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})
		panicChan := make(chan interface{}, 1)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()

			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			// Request completed normally
		case p := <-panicChan:
			// Request panicked
			panic(p)
		case <-ctx.Done():
			// Request timed out
			c.Header("Connection", "close")
			c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
				"error":   "Request timeout",
				"timeout": timeout.String(),
			})
		}
	}
}
