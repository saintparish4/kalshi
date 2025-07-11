package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMetricsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test that MetricsHandler returns a valid gin.HandlerFunc
	handler := MetricsHandler()
	assert.NotNil(t, handler)

	// Test the handler function
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a test request
	req := httptest.NewRequest("GET", "/metrics", nil)
	c.Request = req

	// Call the handler
	handler(c)

	// Check that the response is not empty (Prometheus metrics)
	assert.NotEmpty(t, w.Body.String())
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
}

func TestNewCustomMetricsHandler(t *testing.T) {
	handler := NewCustomMetricsHandler()
	assert.NotNil(t, handler)
}

func TestCustomMetricsHandler_GetMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewCustomMetricsHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.GetMetrics(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check expected fields
	expectedFields := []string{"metrics_endpoint", "format", "available_metrics"}
	for _, field := range expectedFields {
		assert.Contains(t, response, field)
	}

	assert.Equal(t, "/metrics", response["metrics_endpoint"])
	assert.Equal(t, "prometheus", response["format"])

	// Check available metrics
	availableMetrics, ok := response["available_metrics"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, availableMetrics, 5)

	expectedMetrics := []string{
		"gateway_requests_total",
		"gateway_request_duration_seconds",
		"gateway_rate_limit_hits_total",
		"gateway_cache_hits_total",
		"gateway_circuit_breaker_state",
	}

	for _, expectedMetric := range expectedMetrics {
		found := false
		for _, metric := range availableMetrics {
			if metric.(string) == expectedMetric {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected metric %s not found", expectedMetric)
	}
}

func TestMetricsHandler_ContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := MetricsHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/metrics", nil)
	c.Request = req

	handler(c)

	// Check that the content type is set correctly for Prometheus metrics
	contentType := w.Header().Get("Content-Type")
	assert.Contains(t, contentType, "text/plain")
}

func TestCustomMetricsHandler_GetMetrics_ResponseStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewCustomMetricsHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.GetMetrics(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify the response structure is exactly as expected
	assert.Equal(t, "/metrics", response["metrics_endpoint"])
	assert.Equal(t, "prometheus", response["format"])

	// Check that available_metrics is a slice of strings
	availableMetrics, ok := response["available_metrics"].([]interface{})
	assert.True(t, ok)
	assert.Greater(t, len(availableMetrics), 0)

	// Verify all metrics are strings
	for _, metric := range availableMetrics {
		_, ok := metric.(string)
		assert.True(t, ok, "All metrics should be strings")
	}
}

func TestMetricsHandler_WithDifferentMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := MetricsHandler()

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET request",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST request",
			method:         "POST",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(tt.method, "/metrics", nil)
			c.Request = req

			handler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.NotEmpty(t, w.Body.String())
		})
	}
}

func TestCustomMetricsHandler_GetMetrics_EmptyResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewCustomMetricsHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.GetMetrics(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Body.String())

	// Verify the response can be parsed as JSON
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}
