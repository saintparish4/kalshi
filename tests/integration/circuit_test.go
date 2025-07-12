package integration

import (
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"kalshi/internal/circuit"
)

// CircuitBreakerTestSuite tests circuit breaker functionality
type CircuitBreakerTestSuite struct {
	suite.Suite
	testConfig *TestConfig
}

func (suite *CircuitBreakerTestSuite) SetupSuite() {
	suite.testConfig = SetupTestEnvironment(suite.T())
}

func (suite *CircuitBreakerTestSuite) TearDownSuite() {
	CleanupTestEnvironment(suite.testConfig)
}

// TestCircuitBreakerStates tests circuit breaker state transitions
func (suite *CircuitBreakerTestSuite) TestCircuitBreakerStates() {
	manager := suite.testConfig.Gateway.GetCircuitManager()

	suite.Run("Initial Closed State", func() {
		breaker := manager.GetBreaker("test-service", 3, 10*time.Second, 2)
		assert.Equal(suite.T(), circuit.StateClosed, breaker.GetState())
	})

	suite.Run("Failure Threshold", func() {
		breaker := manager.GetBreaker("failing-service", 2, 5*time.Second, 1)

		// Simulate failures
		err1 := breaker.Call(func() error {
			return assert.AnError
		})
		assert.Error(suite.T(), err1)
		assert.Equal(suite.T(), circuit.StateClosed, breaker.GetState())

		err2 := breaker.Call(func() error {
			return assert.AnError
		})
		assert.Error(suite.T(), err2)
		assert.Equal(suite.T(), circuit.StateOpen, breaker.GetState())
	})

	suite.Run("Circuit Breaker Open", func() {
		breaker := manager.GetBreaker("open-service", 1, 5*time.Second, 1)

		// Trigger failure to open circuit
		err := breaker.Call(func() error {
			return assert.AnError
		})
		assert.Error(suite.T(), err)
		assert.Equal(suite.T(), circuit.StateOpen, breaker.GetState())

		// Should reject calls when open
		err = breaker.Call(func() error {
			return nil
		})
		assert.Equal(suite.T(), circuit.ErrCircuitBreakerOpen, err)
	})
}


