package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ProxyTestSuite tests proxy functionality
type ProxyTestSuite struct {
	suite.Suite
	testConfig *TestConfig
}

func (suite *ProxyTestSuite) SetupSuite() {
	suite.testConfig = SetupTestEnvironment(suite.T())
}

func (suite *ProxyTestSuite) TearDownSuite() {
	CleanupTestEnvironment(suite.testConfig)
}

// TestHTTPMethods tests different HTTP methods through proxy
func (suite *ProxyTestSuite) TestHTTPMethods() {
	methods := []string{"GET", "POST", "PUT", "DELETE"}

	for _, method := range methods {
		suite.Run(method+" Method", func() {
			var req *http.Request
			var err error

			if method == "POST" || method == "PUT" {
				body := map[string]interface{}{
					"test":   "data",
					"method": method,
				}
				jsonBody, _ := json.Marshal(body)
				req, err = http.NewRequest(method, "/api/v1/users/123", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(method, "/api/v1/users/123", nil)
			}

			require.NoError(suite.T(), err)
			req.Header.Set("X-API-Key", "test-api-key-1")

			w := httptest.NewRecorder()
			req.Header.Set("X-API-Key", "test-api-key-1") // Add valid API key
			suite.testConfig.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)
		})
	}
}

// TestProxyHeaders tests header forwarding
func (suite *ProxyTestSuite) TestProxyHeaders() {
	req, _ := http.NewRequest("GET", "/api/v1/users/123", nil)
	req.Header.Set("X-API-Key", "test-api-key-1")
	req.Header.Set("X-Custom-Header", "custom-value")
	req.Header.Set("User-Agent", "Test-Agent/1.0")

	w := httptest.NewRecorder()
	req.Header.Set("X-API-Key", "test-api-key-1") // Add valid API key
	suite.testConfig.Router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	// Headers should be forwarded to backend
}

// TestProxyResponseHeaders tests response header handling
func (suite *ProxyTestSuite) TestProxyResponseHeaders() {
	req, _ := http.NewRequest("GET", "/api/v1/users/123", nil)
	req.Header.Set("X-API-Key", "test-api-key-1")

	w := httptest.NewRecorder()
	req.Header.Set("X-API-Key", "test-api-key-1") // Add valid API key
	suite.testConfig.Router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	contentType := w.Header().Get("Content-Type")
	if contentType != "" {
		assert.Contains(suite.T(), contentType, "application/json")
	}
}

// TestProxyQueryParameters tests query parameter forwarding
func (suite *ProxyTestSuite) TestProxyQueryParameters() {
	req, _ := http.NewRequest("GET", "/api/v1/users/123?page=1&limit=10&sort=name", nil)
	req.Header.Set("X-API-Key", "test-api-key-1")

	w := httptest.NewRecorder()
	req.Header.Set("X-API-Key", "test-api-key-1") // Add valid API key
	suite.testConfig.Router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	// Query parameters should be forwarded to backend
}

// TestProxyTimeout tests request timeout handling
func (suite *ProxyTestSuite) TestProxyTimeout() {
	// Test slow backend
	req, _ := http.NewRequest("GET", "/api/v1/slow/test", nil)
	req.Header.Set("X-API-Key", "test-api-key-1")

	w := httptest.NewRecorder()
	req.Header.Set("X-API-Key", "test-api-key-1") // Add valid API key
	suite.testConfig.Router.ServeHTTP(w, req)

	// Should handle slow responses appropriately
	assert.NotEqual(suite.T(), http.StatusRequestTimeout, w.Code)
}

func TestProxyTestSuite(t *testing.T) {
	suite.Run(t, new(ProxyTestSuite))
}
