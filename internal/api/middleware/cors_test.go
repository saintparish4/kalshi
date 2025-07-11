package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		origin         string
		method         string
		expectedStatus int
		expectedOrigin string
	}{
		{
			name:           "Allow all origins",
			origin:         "https://example.com",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedOrigin: "*",
		},
		{
			name:           "Preflight request",
			origin:         "https://example.com",
			method:         "OPTIONS",
			expectedStatus: http.StatusNoContent,
			expectedOrigin: "*",
		},
		{
			name:           "No origin header",
			origin:         "",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedOrigin: "*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(CORS())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedOrigin, w.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, DefaultAllowedMethods, w.Header().Get("Access-Control-Allow-Methods"))
			assert.Equal(t, DefaultAllowedHeaders, w.Header().Get("Access-Control-Allow-Headers"))
			assert.Equal(t, DefaultExposeHeaders, w.Header().Get("Access-Control-Expose-Headers"))
			assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
			assert.Equal(t, CORSMaxAge, w.Header().Get("Access-Control-Max-Age"))
		})
	}
}

func TestCORSWithConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		origin         string
		config         CORSConfig
		expectedStatus int
		expectedOrigin string
	}{
		{
			name:   "Allow all origins",
			origin: "https://example.com",
			config: CORSConfig{
				AllowAllOrigins: true,
			},
			expectedStatus: http.StatusOK,
			expectedOrigin: "*",
		},
		{
			name:   "Restricted origins - allowed",
			origin: "https://allowed.com",
			config: CORSConfig{
				AllowAllOrigins: false,
				AllowedOrigins:  []string{"https://allowed.com", "https://another.com"},
			},
			expectedStatus: http.StatusOK,
			expectedOrigin: "https://allowed.com",
		},
		{
			name:   "Restricted origins - not allowed",
			origin: "https://notallowed.com",
			config: CORSConfig{
				AllowAllOrigins: false,
				AllowedOrigins:  []string{"https://allowed.com", "https://another.com"},
			},
			expectedStatus: http.StatusOK,
			expectedOrigin: "", // No origin header set
		},
		{
			name:   "No origin header with restricted config",
			origin: "",
			config: CORSConfig{
				AllowAllOrigins: false,
				AllowedOrigins:  []string{"https://allowed.com"},
			},
			expectedStatus: http.StatusOK,
			expectedOrigin: "", // No origin header set
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(CORSWithConfig(tt.config))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedOrigin, w.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}

func TestRestrictedCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	allowedOrigins := []string{"https://allowed1.com", "https://allowed2.com"}

	tests := []struct {
		name           string
		origin         string
		expectedStatus int
		expectedOrigin string
	}{
		{
			name:           "Allowed origin 1",
			origin:         "https://allowed1.com",
			expectedStatus: http.StatusOK,
			expectedOrigin: "https://allowed1.com",
		},
		{
			name:           "Allowed origin 2",
			origin:         "https://allowed2.com",
			expectedStatus: http.StatusOK,
			expectedOrigin: "https://allowed2.com",
		},
		{
			name:           "Not allowed origin",
			origin:         "https://notallowed.com",
			expectedStatus: http.StatusOK,
			expectedOrigin: "", // No origin header set
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(RestrictedCORS(allowedOrigins))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedOrigin, w.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}

func TestCORSHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check all CORS headers are set correctly
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, DefaultAllowedMethods, w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, DefaultAllowedHeaders, w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, DefaultExposeHeaders, w.Header().Get("Access-Control-Expose-Headers"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, CORSMaxAge, w.Header().Get("Access-Control-Max-Age"))
}

func TestCORSPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Preflight should return 204 No Content
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Check CORS headers are still set
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, DefaultAllowedMethods, w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, DefaultAllowedHeaders, w.Header().Get("Access-Control-Allow-Headers"))
}
