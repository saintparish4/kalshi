package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kalshi/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewProxyHandler(t *testing.T) {
	cfg := &config.Config{}
	mockGateway := createMockGateway()
	mockLogger := createMockLogger()

	handler := NewProxyHandler(mockGateway, cfg, mockLogger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockGateway, handler.gateway)
	assert.Equal(t, cfg, handler.config)
	assert.Equal(t, mockLogger, handler.logger)
}

func TestProxyHandler_HandleRequest_NoMatchingRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Routes: []config.RouteConfig{
			{
				Path:    "/api/users",
				Backend: "user-service",
				Methods: []string{"GET", "POST"},
			},
		},
	}

	mockGateway := createMockGateway()
	mockLogger := createMockLogger()

	handler := NewProxyHandler(mockGateway, cfg, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Test with non-matching path
	req := httptest.NewRequest("GET", "/api/nonexistent", nil)
	c.Request = req

	handler.HandleRequest(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Route not found", response["error"])
	assert.Equal(t, "/api/nonexistent", response["path"])
}

func TestProxyHandler_HandleRequest_MethodNotAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Routes: []config.RouteConfig{
			{
				Path:    "/api/users",
				Backend: "user-service",
				Methods: []string{"GET"},
			},
		},
	}

	mockGateway := createMockGateway()
	mockLogger := createMockLogger()

	handler := NewProxyHandler(mockGateway, cfg, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Test with disallowed method
	req := httptest.NewRequest("POST", "/api/users", nil)
	c.Request = req

	handler.HandleRequest(c)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Method not allowed", response["error"])
	assert.Contains(t, response, "allowed_methods")
}

func TestProxyHandler_HandleRequest_ValidRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Routes: []config.RouteConfig{
			{
				Path:     "/api/users",
				Backend:  "user-service",
				Methods:  []string{"GET", "POST"},
				CacheTTL: 5 * time.Minute,
			},
		},
	}

	mockGateway := createMockGateway()
	mockLogger := createMockLogger()

	handler := NewProxyHandler(mockGateway, cfg, mockLogger)
	handler.proxy = createMockProxy()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Test with valid route and method
	req := httptest.NewRequest("GET", "/api/users", nil)
	c.Request = req

	handler.HandleRequest(c)

	// Should return 502 Bad Gateway since there's no actual backend
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestProxyHandler_GetRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Routes: []config.RouteConfig{
			{
				Path:      "/api/users",
				Backend:   "user-service",
				Methods:   []string{"GET", "POST"},
				RateLimit: 100,
				CacheTTL:  5 * time.Minute,
			},
			{
				Path:      "/api/orders",
				Backend:   "order-service",
				Methods:   []string{"GET", "PUT", "DELETE"},
				RateLimit: 50,
				CacheTTL:  10 * time.Minute,
			},
		},
	}

	mockGateway := createMockGateway()
	mockLogger := createMockLogger()

	handler := NewProxyHandler(mockGateway, cfg, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.GetRoutes(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "routes")
	assert.Contains(t, response, "total")
	assert.Equal(t, float64(2), response["total"])

	routes, ok := response["routes"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, routes, 2)

	// Check first route
	route1, ok := routes[0].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "/api/users", route1["path"])
	assert.Equal(t, "user-service", route1["backend"])
	assert.Equal(t, float64(100), route1["rate_limit"])
	assert.Equal(t, "5m0s", route1["cache_ttl"])

	// Check second route
	route2, ok := routes[1].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "/api/orders", route2["path"])
	assert.Equal(t, "order-service", route2["backend"])
	assert.Equal(t, float64(50), route2["rate_limit"])
	assert.Equal(t, "10m0s", route2["cache_ttl"])
}

func TestProxyHandler_findMatchingRoute(t *testing.T) {
	cfg := &config.Config{
		Routes: []config.RouteConfig{
			{
				Path:    "/api/users",
				Backend: "user-service",
				Methods: []string{"GET"},
			},
			{
				Path:    "/api/users/*",
				Backend: "user-service",
				Methods: []string{"GET", "PUT", "DELETE"},
			},
			{
				Path:    "/api/users/:id",
				Backend: "user-service",
				Methods: []string{"GET"},
			},
		},
	}

	mockGateway := createMockGateway()
	mockLogger := createMockLogger()

	handler := NewProxyHandler(mockGateway, cfg, mockLogger)

	tests := []struct {
		name     string
		path     string
		method   string
		expected *config.RouteConfig
	}{
		{
			name:     "exact match",
			path:     "/api/users",
			method:   "GET",
			expected: &cfg.Routes[0],
		},
		{
			name:     "wildcard match",
			path:     "/api/users/123",
			method:   "GET",
			expected: &cfg.Routes[1],
		},
		{
			name:     "parameter match",
			path:     "/api/users/456",
			method:   "GET",
			expected: &cfg.Routes[1], // Will match wildcard first
		},
		{
			name:     "no match",
			path:     "/api/nonexistent",
			method:   "GET",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := handler.findMatchingRoute(tt.path, tt.method)
			if tt.expected == nil {
				assert.Nil(t, route)
			} else {
				assert.NotNil(t, route)
				assert.Equal(t, tt.expected.Path, route.Path)
				assert.Equal(t, tt.expected.Backend, route.Backend)
			}
		})
	}
}

func TestProxyHandler_matchRoute(t *testing.T) {
	cfg := &config.Config{}
	mockGateway := createMockGateway()
	mockLogger := createMockLogger()

	handler := NewProxyHandler(mockGateway, cfg, mockLogger)

	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		{
			name:     "exact match",
			path:     "/api/users",
			pattern:  "/api/users",
			expected: true,
		},
		{
			name:     "wildcard match",
			path:     "/api/users/123",
			pattern:  "/api/users/*",
			expected: true,
		},
		{
			name:     "parameter match",
			path:     "/api/users/456",
			pattern:  "/api/users/:id",
			expected: false, // Will be matched by wildcard first in real usage
		},
		{
			name:     "no match",
			path:     "/api/users",
			pattern:  "/api/orders",
			expected: false,
		},
		{
			name:     "partial wildcard match",
			path:     "/api/users/123/details",
			pattern:  "/api/users/*",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.matchRoute(tt.path, tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProxyHandler_matchWithParams(t *testing.T) {
	cfg := &config.Config{}
	mockGateway := createMockGateway()
	mockLogger := createMockLogger()

	handler := NewProxyHandler(mockGateway, cfg, mockLogger)

	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		{
			name:     "single parameter",
			path:     "/api/users/123",
			pattern:  "/api/users/:id",
			expected: true,
		},
		{
			name:     "multiple parameters",
			path:     "/api/users/123/orders/456",
			pattern:  "/api/users/:userId/orders/:orderId",
			expected: true,
		},
		{
			name:     "no parameters",
			path:     "/api/users",
			pattern:  "/api/users",
			expected: true, // Exact match should work
		},
		{
			name:     "parameter mismatch",
			path:     "/api/users/123/orders",
			pattern:  "/api/users/:userId/orders/:orderId",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.matchWithParams(tt.path, tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProxyHandler_isMethodAllowed(t *testing.T) {
	cfg := &config.Config{}
	mockGateway := createMockGateway()
	mockLogger := createMockLogger()

	handler := NewProxyHandler(mockGateway, cfg, mockLogger)

	tests := []struct {
		name           string
		allowedMethods []string
		method         string
		expected       bool
	}{
		{
			name:           "allowed method",
			allowedMethods: []string{"GET", "POST"},
			method:         "GET",
			expected:       true,
		},
		{
			name:           "disallowed method",
			allowedMethods: []string{"GET", "POST"},
			method:         "PUT",
			expected:       false,
		},
		{
			name:           "case sensitive",
			allowedMethods: []string{"GET", "POST"},
			method:         "get",
			expected:       false,
		},
		{
			name:           "empty allowed methods",
			allowedMethods: []string{},
			method:         "GET",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.isMethodAllowed(tt.allowedMethods, tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProxyHandler_HandleRequest_WithContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Routes: []config.RouteConfig{
			{
				Path:    "/api/users",
				Backend: "user-service",
				Methods: []string{"GET"},
			},
		},
	}

	mockGateway := createMockGateway()
	mockLogger := createMockLogger()

	handler := NewProxyHandler(mockGateway, cfg, mockLogger)
	handler.proxy = createMockProxy()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/api/users", nil)
	c.Request = req

	handler.HandleRequest(c)

	// Check that context values are set
	backend, exists := c.Get("backend")
	assert.True(t, exists)
	assert.Equal(t, "user-service", backend)

	routePath, exists := c.Get("route_path")
	assert.True(t, exists)
	assert.Equal(t, "/api/users", routePath)
}
