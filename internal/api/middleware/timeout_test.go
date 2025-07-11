package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		timeout        time.Duration
		handlerDelay   time.Duration
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Request completes within timeout",
			timeout:        100 * time.Millisecond,
			handlerDelay:   50 * time.Millisecond,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"success"}`,
		},
		{
			name:           "Request times out",
			timeout:        100 * time.Millisecond,
			handlerDelay:   200 * time.Millisecond,
			expectedStatus: http.StatusRequestTimeout,
			expectedBody:   `{"error":"Request timeout","timeout":"100ms"}`,
		},
		{
			name:           "Timeout below minimum",
			timeout:        50 * time.Millisecond, // Below MinTimeout
			handlerDelay:   10 * time.Millisecond,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"success"}`,
		},
		{
			name:           "Timeout above maximum",
			timeout:        10 * time.Minute, // Above MaxTimeout
			handlerDelay:   50 * time.Millisecond,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(Timeout(tt.timeout))
			router.GET("/test", func(c *gin.Context) {
				time.Sleep(tt.handlerDelay)
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			} else {
				assert.Empty(t, w.Body.String())
			}
		})
	}
}

func TestTimeoutWithPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Timeout(100 * time.Millisecond))
	router.GET("/panic", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond)
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// This should panic, but the test should not crash
	assert.Panics(t, func() {
		router.ServeHTTP(w, req)
	})
}

func TestTimeoutHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Timeout(100 * time.Millisecond))
	router.GET("/timeout", func(c *gin.Context) {
		time.Sleep(200 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/timeout", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestTimeout, w.Code)
	assert.JSONEq(t, `{"error":"Request timeout","timeout":"100ms"}`, w.Body.String())

	// Check that Connection header is set to close
	assert.Equal(t, "close", w.Header().Get("Connection"))
}

func TestTimeoutValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		timeout        time.Duration
		expectedStatus int
	}{
		{
			name:           "Zero timeout (should use minimum)",
			timeout:        0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Negative timeout (should use minimum)",
			timeout:        -1 * time.Second,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Very large timeout (should use maximum)",
			timeout:        1 * time.Hour,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(Timeout(tt.timeout))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, `{"status":"success"}`, w.Body.String())
		})
	}
}

func TestTimeoutWithDifferentMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Timeout(100 * time.Millisecond))
	router.POST("/test", func(c *gin.Context) {
		time.Sleep(50 * time.Millisecond)
		c.JSON(http.StatusCreated, gin.H{"status": "created"})
	})
	router.PUT("/test", func(c *gin.Context) {
		time.Sleep(50 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	})
	router.DELETE("/test", func(c *gin.Context) {
		time.Sleep(50 * time.Millisecond)
		c.Status(http.StatusNoContent)
	})

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "POST request",
			method:         "POST",
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"status":"created"}`,
		},
		{
			name:           "PUT request",
			method:         "PUT",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"updated"}`,
		},
		{
			name:           "DELETE request",
			method:         "DELETE",
			expectedStatus: http.StatusNoContent,
			expectedBody:   ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			} else {
				assert.Empty(t, w.Body.String())
			}
		})
	}
}

func TestTimeoutWithContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Timeout(100 * time.Millisecond))
	router.GET("/test", func(c *gin.Context) {
		// Check that context has timeout
		select {
		case <-c.Request.Context().Done():
			// Context was cancelled
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "context cancelled"})
		default:
			time.Sleep(50 * time.Millisecond)
			c.JSON(http.StatusOK, gin.H{"status": "success"})
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"success"}`, w.Body.String())
}

func TestTimeoutEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		timeout        time.Duration
		handlerDelay   time.Duration
		expectedStatus int
	}{
		{
			name:           "Exact timeout boundary",
			timeout:        100 * time.Millisecond,
			handlerDelay:   100 * time.Millisecond,
			expectedStatus: http.StatusOK, // At boundary, it might succeed due to timing variance
		},
		{
			name:           "Just under timeout",
			timeout:        100 * time.Millisecond,
			handlerDelay:   99 * time.Millisecond,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Just over timeout",
			timeout:        100 * time.Millisecond,
			handlerDelay:   101 * time.Millisecond,
			expectedStatus: http.StatusRequestTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(Timeout(tt.timeout))
			router.GET("/test", func(c *gin.Context) {
				time.Sleep(tt.handlerDelay)
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
