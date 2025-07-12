package integration

import (
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// GatewayIntegrationTestSuite tests the complete gateway functionality
type GatewayIntegrationTestSuite struct {
	suite.Suite
	testConfig *TestConfig
}

// SetupSuite initializes the test suite
func (suite *GatewayIntegrationTestSuite) SetupSuite() {
	suite.testConfig = SetupTestEnvironment(suite.T())
}

// TearDownSuite cleans up after all tests
func (suite *GatewayIntegrationTestSuite) TearDownSuite() {
	CleanupTestEnvironment(suite.testConfig)
}

// TestHealthEndpoint tests the health check functionality
func (suite *GatewayIntegrationTestSuite) TestHealthEndpoint() {
	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Basic Health Check",
			endpoint:       "/health",
			expectedStatus: http.StatusOK,
			expectedBody:   "healthy",
		},
		{
			name:           "Kubernetes Health Check",
			endpoint:       "/healthz",
			expectedStatus: http.StatusOK,
			expectedBody:   "healthy",
		},
		{
			name:           "Readiness Check",
			endpoint:       "/ready",
			expectedStatus: http.StatusOK,
			expectedBody:   "ready",
		},
		{
			name:           "Liveness Check",
			endpoint:       "/live",
			expectedStatus: http.StatusOK,
			expectedBody:   "alive",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req, err := http.NewRequest("GET", tt.endpoint, nil)
			require.NoError(suite.T(), err)

			w := httptest.NewRecorder()
			suite.testConfig.Router.ServeHTTP(w, req)

			assert.Equal(suite.T(), tt.expectedStatus, w.Code)
			assert.Contains(suite.T(), w.Body.String(), tt.expectedBody)
		})
	}
}

// TestProxyFunctionality tests basic proxy functionality
func (suite *GatewayIntegrationTestSuite) TestProxyFunctionality() {
	// Test successful proxy request
	req, err := http.NewRequest("GET", "/api/v1/users/123", nil)
	require.NoError(suite.T(), err)
	req.Header.Set("X-API-Key", "test-api-key-1")

	w := httptest.NewRecorder()
	suite.testConfig.Router.ServeHTTP(w, req)

	// Should proxy to backend successfully
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestAuthenticationFlow tests authentication mechanisms
func (suite *GatewayIntegrationTestSuite) TestAuthenticationFlow() {
	suite.Run("API Key Authentication", func() {
		// Test without API key
		req, _ := http.NewRequest("GET", "/api/v1/users/123", nil)
		w := httptest.NewRecorder()
		suite.testConfig.Router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

		// Test with valid API key
		req.Header.Set("X-API-Key", "test-api-key-1")
		w = httptest.NewRecorder()
		suite.testConfig.Router.ServeHTTP(w, req)
		assert.NotEqual(suite.T(), http.StatusUnauthorized, w.Code)
	})

	suite.Run("JWT Authentication", func() {
		// Generate test token
		token, err := GenerateTestJWT(suite.testConfig.JWTManager, "test-user", "user")
		require.NoError(suite.T(), err)

		req, _ := http.NewRequest("GET", "/api/v1/users/456", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		suite.testConfig.Router.ServeHTTP(w, req)
		assert.NotEqual(suite.T(), http.StatusUnauthorized, w.Code)
	})
}

// TestErrorHandling tests error scenarios
func (suite *GatewayIntegrationTestSuite) TestErrorHandling() {
	suite.Run("Route Not Found", func() {
		req, _ := http.NewRequest("GET", "/nonexistent", nil)
		w := httptest.NewRecorder()
		suite.testConfig.Router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	})

	suite.Run("Method Not Allowed", func() {
		req, _ := http.NewRequest("PATCH", "/api/v1/users/123", nil)
		req.Header.Set("X-API-Key", "test-api-key-1")
		w := httptest.NewRecorder()
		suite.testConfig.Router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusMethodNotAllowed, w.Code)
	})
}

// Run the gateway integration test suite
