package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"kalshi/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestLogging(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestLogging(log))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"success"}`, w.Body.String())
}

func TestRequestLoggingWithContext(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestLogging(log))
	router.GET("/test", func(c *gin.Context) {
		// Set context values that should be logged
		c.Set(LogContextUserID, "user123")
		c.Set(LogContextAuthMethod, "jwt")
		c.Set(LogContextBackend, "api-server")
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"success"}`, w.Body.String())
}

func TestRequestLoggingError(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestLogging(log))
	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
	})

	req := httptest.NewRequest("GET", "/error", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"server error"}`, w.Body.String())
}

func TestStructuredLogging(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(StructuredLogging(log))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/test?param=value", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"success"}`, w.Body.String())
}

func TestStructuredLoggingWithContext(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(StructuredLogging(log))
	router.GET("/test", func(c *gin.Context) {
		// Set context values that should be logged
		c.Set(LogContextUserID, "user456")
		c.Set(LogContextAuthMethod, "api_key")
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"success"}`, w.Body.String())
}

func TestStructuredLoggingError(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(StructuredLogging(log))
	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client error"})
	})

	req := httptest.NewRequest("GET", "/error", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"client error"}`, w.Body.String())
}

func TestStructuredLoggingWithErrors(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(StructuredLogging(log))
	router.GET("/test", func(c *gin.Context) {
		c.Error(assert.AnError)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"success"}`, w.Body.String())
}

func TestLoggingDefaultValues(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(StructuredLogging(log))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"success"}`, w.Body.String())
}

func TestLoggingWithQueryParams(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(StructuredLogging(log))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/test?param1=value1&param2=value2", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"success"}`, w.Body.String())
}

func TestLoggingWithDifferentMethods(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(StructuredLogging(log))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"status": "created"})
	})
	router.PUT("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	})
	router.DELETE("/test", func(c *gin.Context) {
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
