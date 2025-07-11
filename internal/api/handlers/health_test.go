package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kalshi/internal/config"
	"kalshi/internal/gateway"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewHealthHandler(t *testing.T) {
	mockGateway := createMockGateway()
	handler := NewHealthHandler(mockGateway)

	assert.NotNil(t, handler)
	assert.Equal(t, mockGateway, handler.gateway)
}

func TestHealthHandler_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGateway := createMockGateway()
	handler := NewHealthHandler(mockGateway)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.Health(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expectedFields := []string{"status", "timestamp", "version", "service"}
	for _, field := range expectedFields {
		assert.Contains(t, response, field)
	}

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.Equal(t, "kalshi-api-gateway", response["service"])

	// Check timestamp is present and valid
	timestamp, ok := response["timestamp"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, timestamp)
}

func TestHealthHandler_DetailedHealth(t *testing.T) {
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
	handler := NewHealthHandler(mockGateway)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.DetailedHealth(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check basic fields
	expectedBasicFields := []string{"status", "timestamp", "version", "service", "uptime"}
	for _, field := range expectedBasicFields {
		assert.Contains(t, response, field)
	}

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.Equal(t, "kalshi-api-gateway", response["service"])

	// Check backend information
	assert.Contains(t, response, "healthy_backends")
	assert.Contains(t, response, "all_backends")
	assert.Contains(t, response, "circuit_breaker")

	healthyBackends, ok := response["healthy_backends"].(map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, healthyBackends, 2)

	allBackends, ok := response["all_backends"].(map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, allBackends, 2)

	// Check circuit breaker information
	circuitBreaker, ok := response["circuit_breaker"].(map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, circuitBreaker, 2)
}

func TestHealthHandler_Readiness(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		healthyBackends int
		expectedStatus  int
		expectedBody    map[string]interface{}
	}{
		{
			name:            "ready with healthy backends",
			healthyBackends: 2,
			expectedStatus:  http.StatusOK,
			expectedBody: map[string]interface{}{
				"status":           "ready",
				"healthy_backends": float64(2),
			},
		},
		{
			name:            "not ready with no healthy backends",
			healthyBackends: 0,
			expectedStatus:  http.StatusServiceUnavailable,
			expectedBody: map[string]interface{}{
				"status": "not ready",
				"reason": "no healthy backends available",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config based on test case
			var backends []config.BackendConfig
			if tt.healthyBackends > 0 {
				backends = []config.BackendConfig{
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
				}
			}

			cfg := &config.Config{Backend: backends}
			mockGateway := gateway.New(cfg, nil, nil)

			handler := NewHealthHandler(mockGateway)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			handler.Readiness(c)

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

func TestHealthHandler_Liveness(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGateway := createMockGateway()
	handler := NewHealthHandler(mockGateway)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.Liveness(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expectedFields := []string{"status", "timestamp"}
	for _, field := range expectedFields {
		assert.Contains(t, response, field)
	}

	assert.Equal(t, "alive", response["status"])

	// Check timestamp is present and valid
	timestamp, ok := response["timestamp"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, timestamp)
}

func TestHealthHandler_DetailedHealth_WithUnhealthyBackends(t *testing.T) {
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

	// Mark one backend as unhealthy
	backends := mockGateway.GetBackendManager().GetAllBackends()
	if len(backends) > 0 {
		backends[0].IsHealthy = false
	}

	handler := NewHealthHandler(mockGateway)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.DetailedHealth(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check that we have both healthy and unhealthy backends
	healthyBackends, ok := response["healthy_backends"].(map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, healthyBackends, 1) // Only one healthy backend

	allBackends, ok := response["all_backends"].(map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, allBackends, 2) // Both backends in all_backends
}

func TestHealthHandler_Readiness_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupBackends  func(*gateway.BackendManager)
		expectedStatus int
	}{
		{
			name: "single healthy backend",
			setupBackends: func(bm *gateway.BackendManager) {
				bm.AddBackend("backend1", "http://backend1:8080", "/health", 1)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "multiple unhealthy backends",
			setupBackends: func(bm *gateway.BackendManager) {
				bm.AddBackend("backend1", "http://backend1:8080", "/health", 1)
				bm.AddBackend("backend2", "http://backend2:8080", "/health", 2)

				// Mark all backends as unhealthy
				backends := bm.GetAllBackends()
				for _, backend := range backends {
					backend.IsHealthy = false
				}
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config based on test case
			cfg := &config.Config{Backend: []config.BackendConfig{}}
			mockGateway := gateway.New(cfg, nil, nil)

			// Setup backends using the gateway's backend manager
			tt.setupBackends(mockGateway.GetBackendManager())

			handler := NewHealthHandler(mockGateway)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			handler.Readiness(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
