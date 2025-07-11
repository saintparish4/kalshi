package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kalshi/internal/config"
	"kalshi/internal/gateway"
	"kalshi/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewAdminHandler(t *testing.T) {
	mockGateway := createMockGateway()
	mockLogger := &logger.Logger{}

	handler := NewAdminHandler(mockGateway, mockLogger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockGateway, handler.gateway)
	assert.Equal(t, mockLogger, handler.logger)
}

func TestAdminHandler_GetBackends(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a gateway with test data
	cfg := &config.Config{
		Backend: []config.BackendConfig{
			{
				Name:        "backend1",
				URL:         "http://backend1:8080",
				HealthCheck: "/health",
				Weight:      1,
			},
			{
				Name:        "backend2",
				URL:         "http://backend2:8080",
				HealthCheck: "/health",
				Weight:      2,
			},
		},
	}

	mockGateway := gateway.New(cfg, nil, nil)
	mockLogger := &logger.Logger{}

	handler := NewAdminHandler(mockGateway, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.GetBackends(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "backends")
	assert.Contains(t, response, "total")
	assert.Equal(t, float64(2), response["total"])

	backends, ok := response["backends"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, backends, 2)
}

func TestAdminHandler_GetCircuitBreakers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a gateway with test data
	cfg := &config.Config{
		Backend: []config.BackendConfig{
			{
				Name:        "backend1",
				URL:         "http://backend1:8080",
				HealthCheck: "/health",
				Weight:      1,
			},
			{
				Name:        "backend2",
				URL:         "http://backend2:8080",
				HealthCheck: "/health",
				Weight:      2,
			},
		},
	}

	mockGateway := createMockGatewayWithBackends(cfg.Backend)
	mockLogger := createMockLogger()

	handler := NewAdminHandler(mockGateway, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.GetCircuitBreakers(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "circuit_breakers")
	assert.Contains(t, response, "total")
	assert.Equal(t, float64(2), response["total"])
}

func TestAdminHandler_ResetCircuitBreaker(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		backend        string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "success",
			backend:        "backend1",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Circuit breaker reset successfully",
				"backend": "backend1",
			},
		},
		{
			name:           "missing backend parameter",
			backend:        "",
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Backend parameter is required",
			},
		},
		{
			name:           "backend not found",
			backend:        "nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "circuit breaker for backend nonexistent not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config based on test case
			var backends []config.BackendConfig
			if tt.backend != "" && tt.backend != "nonexistent" {
				backends = []config.BackendConfig{
					{
						Name:        tt.backend,
						URL:         "http://" + tt.backend + ":8080",
						HealthCheck: "/health",
						Weight:      1,
					},
				}
			}

			mockGateway := createMockGatewayWithBackends(backends)
			mockLogger := createMockLogger()

			handler := NewAdminHandler(mockGateway, mockLogger)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set up the request with parameters
			req := httptest.NewRequest("POST", "/admin/circuit-breaker/"+tt.backend, nil)
			c.Request = req
			c.Params = gin.Params{{Key: "backend", Value: tt.backend}}

			handler.ResetCircuitBreaker(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for key, expectedValue := range tt.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}
		})
	}
}

func TestAdminHandler_GetStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a gateway with test data
	cfg := &config.Config{
		Backend: []config.BackendConfig{
			{
				Name:        "backend1",
				URL:         "http://backend1:8080",
				HealthCheck: "/health",
				Weight:      1,
			},
			{
				Name:        "backend2",
				URL:         "http://backend2:8080",
				HealthCheck: "/health",
				Weight:      2,
			},
			{
				Name:        "backend3",
				URL:         "http://backend3:8080",
				HealthCheck: "/health",
				Weight:      3,
			},
		},
	}

	mockGateway := createMockGatewayWithBackends(cfg.Backend)
	mockLogger := createMockLogger()

	handler := NewAdminHandler(mockGateway, mockLogger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.GetStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expectedStats := map[string]interface{}{
		"total_backends":     float64(3),
		"healthy_backends":   float64(3), // All backends start as healthy
		"unhealthy_backends": float64(0),
		"total_circuits":     float64(3),
		"open_circuits":      float64(0), // All circuits start closed
	}

	for key, expectedValue := range expectedStats {
		assert.Equal(t, expectedValue, response[key])
	}

	assert.Contains(t, response, "uptime")
}
