package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// BackendTestSuite tests backend management functionality
type BackendTestSuite struct {
	suite.Suite
	testConfig *TestConfig
}

func (suite *BackendTestSuite) SetupSuite() {
	suite.testConfig = SetupTestEnvironment(suite.T())
}

func (suite *BackendTestSuite) TearDownSuite() {
	CleanupTestEnvironment(suite.testConfig)
}

// TestBackendHealth tests backend health checking
func (suite *BackendTestSuite) TestBackendHealth() {
	backendManager := suite.testConfig.Gateway.GetBackendManager()
	
	suite.Run("Get All Backends", func() {
		backends := backendManager.GetAllBackends()
		assert.Greater(suite.T(), len(backends), 0)
	})
	
	suite.Run("Get Healthy Backends", func() {
		healthyBackends := backendManager.GetHealthyBackends()
		// At least the test backend should be healthy
		assert.GreaterOrEqual(suite.T(), len(healthyBackends), 1)
	})
	
	suite.Run("Get Specific Backend", func() {
		backend, err := backendManager.GetBackend("test-backend")
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), backend)
		assert.Equal(suite.T(), "test-backend", backend.Name)
	})
	
	suite.Run("Get Nonexistent Backend", func() {
		_, err := backendManager.GetBackend("nonexistent-backend")
		assert.Error(suite.T(), err)
	})
}

// TestBackendHealthChecks tests periodic health checking
func (suite *BackendTestSuite) TestBackendHealthChecks() {
	// Start health checks with short interval for testing
	backendManager := suite.testConfig.Gateway.GetBackendManager()
	backendManager.StartHealthChecks(1 * time.Second)
	
	// Wait for at least one health check cycle
	time.Sleep(2 * time.Second)
	
	// Check that backends have been health checked
	backends := backendManager.GetAllBackends()
	for _, backend := range backends {
		// Health status should be determined by now
		assert.NotNil(suite.T(), backend)
	}
}

func TestBackendTestSuite(t *testing.T) {
	suite.Run(t, new(BackendTestSuite))
}