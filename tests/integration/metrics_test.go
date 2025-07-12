package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// MetricsTestSuite tests metrics functionality
type MetricsTestSuite struct {
	suite.Suite
	testConfig *TestConfig
}

func (suite *MetricsTestSuite) SetupSuite() {
	suite.testConfig = SetupTestEnvironment(suite.T())
}

func (suite *MetricsTestSuite) TearDownSuite() {
	CleanupTestEnvironment(suite.testConfig)
}

// TestMetricsEndpoint tests the Prometheus metrics endpoint
func (suite *MetricsTestSuite) TestMetricsEndpoint() {
	req, _ := http.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	suite.testConfig.Router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Check for Prometheus metrics format
	body := w.Body.String()
	assert.Contains(suite.T(), body, "# HELP")
	assert.Contains(suite.T(), body, "# TYPE")
}

// TestMetricsCollection tests that metrics are being collected
func (suite *MetricsTestSuite) TestMetricsCollection() {
	// Make a request that should generate metrics
	req, _ := http.NewRequest("GET", "/api/v1/users/123", nil)
	req.Header.Set("X-API-Key", "test-api-key-1")
	w := httptest.NewRecorder()
	suite.testConfig.Router.ServeHTTP(w, req)

	// Check metrics endpoint
	metricsReq, _ := http.NewRequest("GET", "/metrics", nil)
	metricsW := httptest.NewRecorder()
	suite.testConfig.Router.ServeHTTP(metricsW, metricsReq)

	assert.Equal(suite.T(), http.StatusOK, metricsW.Code)

	metricsBody := metricsW.Body.String()
	// Should contain request metrics
	assert.Contains(suite.T(), metricsBody, "http_requests_total")
}

func TestMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(MetricsTestSuite))
}
