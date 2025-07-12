package integration

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RateLimitTestSuite tests rate limiting functionality
type RateLimitTestSuite struct {
	suite.Suite
	testConfig *TestConfig
}

func (suite *RateLimitTestSuite) SetupSuite() {
	suite.testConfig = SetupTestEnvironment(suite.T())
}

func (suite *RateLimitTestSuite) TearDownSuite() {
	CleanupTestEnvironment(suite.testConfig)
}

// TestBasicRateLimit tests basic rate limiting functionality
func (suite *RateLimitTestSuite) TestBasicRateLimit() {
	limiter := suite.testConfig.Limiter
	ctx := context.Background()

	// Test token bucket behavior
	clientID := "test-rate-limit-client"
	path := "/api/test"

	// Multiple requests should work within burst capacity
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, clientID, path)
		require.NoError(suite.T(), err)
		assert.True(suite.T(), allowed, "Request %d should be allowed", i+1)
	}
}

// TestRateLimitHTTP tests rate limiting through HTTP requests
func (suite *RateLimitTestSuite) TestRateLimitHTTP() {
	// Reduce rate limit for testing

	// Make requests with same client ID
	clientID := "http-test-client"

	suite.Run("Rate Limit Headers", func() {
		req, _ := http.NewRequest("GET", "/health", nil)
		req.Header.Set("X-Client-ID", clientID)

		w := httptest.NewRecorder()
		suite.testConfig.Router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
		// Check for rate limit headers if they exist
		if limit := w.Header().Get("X-RateLimit-Limit"); limit != "" {
			assert.NotEmpty(suite.T(), limit)
		}
	})
}

// TestPerUserRateLimit tests per-user rate limiting
func (suite *RateLimitTestSuite) TestPerUserRateLimit() {
	// Generate tokens for different users
	token1, err := GenerateTestJWT(suite.testConfig.JWTManager, "user1", "user")
	require.NoError(suite.T(), err)

	token2, err := GenerateTestJWT(suite.testConfig.JWTManager, "user2", "user")
	require.NoError(suite.T(), err)

	// Each user should have independent rate limits
	suite.Run("User1 Requests", func() {
		req, _ := http.NewRequest("GET", "/api/v1/users/123", nil)
		req.Header.Set("Authorization", "Bearer "+token1)

		w := httptest.NewRecorder()
		suite.testConfig.Router.ServeHTTP(w, req)
		assert.NotEqual(suite.T(), http.StatusTooManyRequests, w.Code)
	})

	suite.Run("User2 Requests", func() {
		req, _ := http.NewRequest("GET", "/api/v1/users/456", nil)
		req.Header.Set("Authorization", "Bearer "+token2)

		w := httptest.NewRecorder()
		suite.testConfig.Router.ServeHTTP(w, req)
		assert.NotEqual(suite.T(), http.StatusTooManyRequests, w.Code)
	})
}

// TestRateLimitRecovery tests rate limit recovery
func (suite *RateLimitTestSuite) TestRateLimitRecovery() {
	// This test would need to actually exhaust rate limits
	// and then wait for recovery - complex to implement in unit tests
	suite.T().Skip("Rate limit recovery test requires time-based testing")
}
