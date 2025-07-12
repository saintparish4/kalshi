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
)

// CacheTestSuite tests caching functionality
type CacheTestSuite struct {
	suite.Suite
	testConfig *TestConfig
}

func (suite *CacheTestSuite) SetupSuite() {
	suite.testConfig = SetupTestEnvironment(suite.T())
}

func (suite *CacheTestSuite) TearDownSuite() {
	CleanupTestEnvironment(suite.testConfig)
}

// TestCacheBasicOperations tests basic cache operations
func (suite *CacheTestSuite) TestCacheBasicOperations() {
	cache := suite.testConfig.CacheManager
	ctx := context.Background()

	suite.Run("Set and Get", func() {
		key := "test-key"
		value := []byte("test-value")

		err := cache.Set(ctx, key, value, time.Minute)
		require.NoError(suite.T(), err)

		retrieved, err := cache.Get(ctx, key)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), value, retrieved)
	})

	suite.Run("Cache Miss", func() {
		_, err := cache.Get(ctx, "nonexistent-key")
		assert.Error(suite.T(), err)
	})

	suite.Run("Delete", func() {
		key := "delete-test-key"
		value := []byte("delete-test-value")

		err := cache.Set(ctx, key, value, time.Minute)
		require.NoError(suite.T(), err)

		err = cache.Delete(ctx, key)
		require.NoError(suite.T(), err)

		_, err = cache.Get(ctx, key)
		assert.Error(suite.T(), err)
	})
}

// TestHTTPCaching tests caching through HTTP requests
func (suite *CacheTestSuite) TestHTTPCaching() {
	// Make a GET request that should be cached
	req, _ := http.NewRequest("GET", "/api/v1/users/123", nil)
	w := httptest.NewRecorder()
	req.Header.Set("X-API-Key", "test-api-key-1") // Add valid API key
	suite.testConfig.Router.ServeHTTP(w, req)

	// Second request (should potentially hit cache)
	w2 := httptest.NewRecorder()
	suite.testConfig.Router.ServeHTTP(w2, req)

	// Both should succeed
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), http.StatusOK, w2.Code)

	// Check for cache headers if implemented
	if cacheHeader := w2.Header().Get("X-Cache"); cacheHeader != "" {
		// Cache header found, verify it indicates a hit
		assert.Contains(suite.T(), cacheHeader, "HIT")
	}
}

// TestCacheExpiration tests cache TTL functionality
func (suite *CacheTestSuite) TestCacheExpiration() {
	cache := suite.testConfig.CacheManager
	ctx := context.Background()

	key := "expiration-test"
	value := []byte("expires-soon")

	// Set with very short TTL
	err := cache.Set(ctx, key, value, 100*time.Millisecond)
	require.NoError(suite.T(), err)

	// Should exist immediately
	retrieved, err := cache.Get(ctx, key)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), value, retrieved)

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Should be expired now
	_, err = cache.Get(ctx, key)
	assert.Error(suite.T(), err)
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}
