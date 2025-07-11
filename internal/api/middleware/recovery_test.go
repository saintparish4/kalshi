package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"kalshi/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRecovery(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Recovery(log))
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})
	router.GET("/normal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Normal request",
			path:           "/normal",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"success"}`,
		},
		{
			name:           "Panic recovery",
			path:           "/panic",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error","request_id":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			req.Header.Set("User-Agent", "test-agent")
			req.RemoteAddr = "127.0.0.1:12345"
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestRecoveryWithRequestID(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID())
	router.Use(Recovery(log))
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	req.Header.Set("X-Request-ID", "test-request-id")
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"Internal server error","request_id":"test-request-id"}`, w.Body.String())
}

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		assert.True(t, exists)
		assert.NotEmpty(t, requestID)
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	tests := []struct {
		name           string
		requestID      string
		expectedStatus int
	}{
		{
			name:           "No request ID header",
			requestID:      "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "With request ID header",
			requestID:      "custom-request-id",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.requestID != "" {
				req.Header.Set("X-Request-ID", tt.requestID)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check that request ID header is set in response
			responseRequestID := w.Header().Get("X-Request-ID")
			assert.NotEmpty(t, responseRequestID)

			// If we provided a custom request ID, it should be used
			if tt.requestID != "" {
				assert.Equal(t, tt.requestID, responseRequestID)
			}
		})
	}
}

func TestRequestIDGeneration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		assert.True(t, exists)
		assert.NotEmpty(t, requestID)
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check that a request ID was generated and set in response header
	responseRequestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, responseRequestID)

	// The generated request ID should be numeric (Unix timestamp)
	assert.Regexp(t, `^\d+$`, responseRequestID)
}

func TestRecoveryWithDifferentPanicTypes(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID())
	router.Use(Recovery(log))
	router.GET("/string-panic", func(c *gin.Context) {
		panic("string panic")
	})
	router.GET("/error-panic", func(c *gin.Context) {
		panic(assert.AnError)
	})
	router.GET("/int-panic", func(c *gin.Context) {
		panic(42)
	})

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "String panic",
			path:           "/string-panic",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error","request_id":""}`,
		},
		{
			name:           "Error panic",
			path:           "/error-panic",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error","request_id":""}`,
		},
		{
			name:           "Int panic",
			path:           "/int-panic",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error","request_id":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			req.Header.Set("User-Agent", "test-agent")
			req.RemoteAddr = "127.0.0.1:12345"
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestRecoveryWithContext(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID())
	router.Use(Recovery(log))
	router.GET("/panic", func(c *gin.Context) {
		// Set some context values before panic
		c.Set("user_id", "user123")
		c.Set("auth_method", "jwt")
		panic("test panic with context")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"Internal server error","request_id":""}`, w.Body.String())
}
