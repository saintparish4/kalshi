package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"kalshi/internal/auth"
)

// AuthTestSuite tests authentication functionality
type AuthTestSuite struct {
	suite.Suite
	testConfig *TestConfig
}

func (suite *AuthTestSuite) SetupSuite() {
	suite.testConfig = SetupTestEnvironment(suite.T())
}

func (suite *AuthTestSuite) TearDownSuite() {
	CleanupTestEnvironment(suite.testConfig)
}

// TestJWTAuthentication tests JWT token validation
func (suite *AuthTestSuite) TestJWTAuthentication() {
	jwtManager := suite.testConfig.JWTManager

	suite.Run("Valid JWT Token", func() {
		tokenPair, err := jwtManager.GenerateTokenPair("test-user", "admin")
		require.NoError(suite.T(), err)

		// Validate the token
		claims, err := jwtManager.ValidateToken(context.Background(), tokenPair.AccessToken)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), "test-user", claims.UserID)
		assert.Equal(suite.T(), "admin", claims.Role)
	})

	suite.Run("Invalid JWT Token", func() {
		invalidToken := "invalid.jwt.token"

		_, err := jwtManager.ValidateToken(context.Background(), invalidToken)
		assert.Error(suite.T(), err)
	})

	suite.Run("Expired JWT Token", func() {
		// Create a JWT manager with very short expiry
		shortJWT := auth.NewJWTManager("test-secret-key-32-chars-minimum-required", 1*time.Millisecond, 1*time.Second)
		tokenPair, err := shortJWT.GenerateTokenPair("test-user", "user")
		require.NoError(suite.T(), err)

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		_, err = shortJWT.ValidateToken(context.Background(), tokenPair.AccessToken)
		assert.Error(suite.T(), err)
	})
}

// TestAPIKeyAuthentication tests API key validation
func (suite *AuthTestSuite) TestAPIKeyAuthentication() {
	suite.Run("Valid API Key via Header", func() {
		req, _ := http.NewRequest("GET", "/api/v1/users/123", nil)
		req.Header.Set("X-API-Key", "test-api-key-1")

		w := httptest.NewRecorder()
		suite.testConfig.Router.ServeHTTP(w, req)

		assert.NotEqual(suite.T(), http.StatusUnauthorized, w.Code)
	})

	suite.Run("Invalid API Key", func() {
		req, _ := http.NewRequest("GET", "/api/v1/users/123", nil)
		req.Header.Set("X-API-Key", "invalid-key")

		w := httptest.NewRecorder()
		suite.testConfig.Router.ServeHTTP(w, req)

		// OptionalAuth allows anonymous access if API key is invalid
		assert.Equal(suite.T(), http.StatusOK, w.Code)
	})

	suite.Run("Missing API Key", func() {
		req, _ := http.NewRequest("GET", "/api/v1/users/123", nil)
		// No API key header

		w := httptest.NewRecorder()
		suite.testConfig.Router.ServeHTTP(w, req)

		// OptionalAuth allows anonymous access if API key is missing
		assert.Equal(suite.T(), http.StatusOK, w.Code)
	})
}

// TestAuthenticationFlow tests complete authentication flows
func (suite *AuthTestSuite) TestAuthenticationFlow() {
	suite.Run("JWT in Authorization Header", func() {
		token, err := GenerateTestJWT(suite.testConfig.JWTManager, "jwt-user", "user")
		require.NoError(suite.T(), err)

		req, _ := http.NewRequest("GET", "/api/v1/users/jwt-test", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		suite.testConfig.Router.ServeHTTP(w, req)

		assert.NotEqual(suite.T(), http.StatusUnauthorized, w.Code)
	})

	suite.Run("Malformed Authorization Header", func() {
		req, _ := http.NewRequest("GET", "/api/v1/users/123", nil)
		req.Header.Set("Authorization", "InvalidFormat token")

		w := httptest.NewRecorder()
		suite.testConfig.Router.ServeHTTP(w, req)

		// OptionalAuth allows anonymous access if JWT is malformed
		assert.Equal(suite.T(), http.StatusOK, w.Code)
	})
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
