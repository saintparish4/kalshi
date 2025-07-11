package gateway

import (
	"testing"
	"time"

	"kalshi/internal/config"
	"kalshi/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Create test dependencies
	cfg := &config.Config{
		Backend: []config.BackendConfig{
			{
				Name:        "test-backend",
				URL:         "http://localhost:8080",
				HealthCheck: "/health",
				Weight:      1,
			},
		},
	}

	cacheManager := &MockCache{}
	logger := &logger.Logger{}

	// Test gateway creation
	gateway := New(cfg, cacheManager, logger)

	// Verify gateway is properly initialized
	assert.NotNil(t, gateway)
	assert.Equal(t, cfg, gateway.config)
	assert.NotNil(t, gateway.backendManager)
	assert.Equal(t, cacheManager, gateway.cacheManager)
	assert.NotNil(t, gateway.circuitManager)
	assert.NotNil(t, gateway.proxy)
	assert.Equal(t, logger, gateway.logger)
}

func TestGateway_GetProxy(t *testing.T) {
	cfg := &config.Config{
		Backend: []config.BackendConfig{
			{
				Name:        "test-backend",
				URL:         "http://localhost:8080",
				HealthCheck: "/health",
				Weight:      1,
			},
		},
	}

	cacheManager := NewMockCache()
	logger := &logger.Logger{}

	gateway := New(cfg, cacheManager, logger)

	// Test GetProxy method
	proxy := gateway.GetProxy()
	assert.NotNil(t, proxy)
	assert.Equal(t, gateway.proxy, proxy)
}

func TestGateway_GetBackendManager(t *testing.T) {
	cfg := &config.Config{
		Backend: []config.BackendConfig{
			{
				Name:        "test-backend",
				URL:         "http://localhost:8080",
				HealthCheck: "/health",
				Weight:      1,
			},
		},
	}

	cacheManager := NewMockCache()
	logger := &logger.Logger{}

	gateway := New(cfg, cacheManager, logger)

	// Test GetBackendManager method
	backendManager := gateway.GetBackendManager()
	assert.NotNil(t, backendManager)
	assert.Equal(t, gateway.backendManager, backendManager)
}

func TestGateway_GetCircuitManager(t *testing.T) {
	cfg := &config.Config{
		Backend: []config.BackendConfig{
			{
				Name:        "test-backend",
				URL:         "http://localhost:8080",
				HealthCheck: "/health",
				Weight:      1,
			},
		},
	}

	cacheManager := NewMockCache()
	logger := &logger.Logger{}

	gateway := New(cfg, cacheManager, logger)

	// Test GetCircuitManager method
	circuitManager := gateway.GetCircuitManager()
	assert.NotNil(t, circuitManager)
	assert.Equal(t, gateway.circuitManager, circuitManager)
}

func TestGateway_initializeBackends(t *testing.T) {
	tests := []struct {
		name        string
		backendCfgs []config.BackendConfig
		expectError bool
	}{
		{
			name: "valid backends",
			backendCfgs: []config.BackendConfig{
				{
					Name:        "backend1",
					URL:         "http://localhost:8080",
					HealthCheck: "/health",
					Weight:      1,
				},
				{
					Name:        "backend2",
					URL:         "http://localhost:8081",
					HealthCheck: "/health",
					Weight:      2,
				},
			},
			expectError: false,
		},
		{
			name: "invalid URL",
			backendCfgs: []config.BackendConfig{
				{
					Name:        "invalid-backend",
					URL:         "invalid-url",
					HealthCheck: "/health",
					Weight:      1,
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Backend: tt.backendCfgs,
			}

			cacheManager := NewMockCache()
			logger := &logger.Logger{}

			gateway := New(cfg, cacheManager, logger)

			// Verify backends were added (for valid cases)
			if !tt.expectError {
				for _, backendCfg := range tt.backendCfgs {
					backend, err := gateway.backendManager.GetBackend(backendCfg.Name)
					require.NoError(t, err)
					assert.Equal(t, backendCfg.Name, backend.Name)
					assert.Equal(t, backendCfg.Weight, backend.Weight)
				}
			}
		})
	}
}

func TestGateway_HealthCheckStart(t *testing.T) {
	cfg := &config.Config{
		Backend: []config.BackendConfig{
			{
				Name:        "test-backend",
				URL:         "http://localhost:8080",
				HealthCheck: "/health",
				Weight:      1,
			},
		},
	}

	cacheManager := NewMockCache()
	logger := &logger.Logger{}

	gateway := New(cfg, cacheManager, logger)

	// Verify that health checks are started
	// We can't directly test the goroutine, but we can verify the backend manager exists
	assert.NotNil(t, gateway.backendManager)

	// Give some time for health checks to potentially start
	time.Sleep(10 * time.Millisecond)
}

func TestGateway_WithEmptyConfig(t *testing.T) {
	// Test with empty config
	cfg := &config.Config{
		Backend: []config.BackendConfig{},
	}

	cacheManager := NewMockCache()
	logger := &logger.Logger{}

	gateway := New(cfg, cacheManager, logger)

	// Should still create gateway successfully
	assert.NotNil(t, gateway)
	assert.NotNil(t, gateway.backendManager)
	assert.NotNil(t, gateway.circuitManager)
	assert.NotNil(t, gateway.proxy)
}

func TestGateway_WithNilDependencies(t *testing.T) {
	// Test with nil dependencies (should handle gracefully)
	cfg := &config.Config{
		Backend: []config.BackendConfig{
			{
				Name:        "test-backend",
				URL:         "http://localhost:8080",
				HealthCheck: "/health",
				Weight:      1,
			},
		},
	}

	// Test with nil cache manager
	gateway := New(cfg, nil, &logger.Logger{})
	assert.NotNil(t, gateway)
	assert.Nil(t, gateway.cacheManager)

	// Test with nil logger
	gateway = New(cfg, NewMockCache(), nil)
	assert.NotNil(t, gateway)
	assert.Nil(t, gateway.logger)
}

func TestGateway_ConcurrentAccess(t *testing.T) {
	cfg := &config.Config{
		Backend: []config.BackendConfig{
			{
				Name:        "test-backend",
				URL:         "http://localhost:8080",
				HealthCheck: "/health",
				Weight:      1,
			},
		},
	}

	cacheManager := NewMockCache()
	logger := &logger.Logger{}

	gateway := New(cfg, cacheManager, logger)

	// Test concurrent access to getter methods
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			proxy := gateway.GetProxy()
			assert.NotNil(t, proxy)

			backendManager := gateway.GetBackendManager()
			assert.NotNil(t, backendManager)

			circuitManager := gateway.GetCircuitManager()
			assert.NotNil(t, circuitManager)

			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
