package testing

import (
	"testing"
	"time"

	"kalshi/internal/config"
	"kalshi/internal/gateway"
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
	gw := gateway.New(cfg, cacheManager, logger)

	// Verify gateway is properly initialized
	assert.NotNil(t, gw)
	assert.Equal(t, cfg, gw.GetConfig())
	assert.NotNil(t, gw.GetBackendManager())
	assert.Equal(t, cacheManager, gw.GetCacheManager())
	assert.NotNil(t, gw.GetCircuitManager())
	assert.NotNil(t, gw.GetProxy())
	assert.Equal(t, logger, gw.GetLogger())
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

	gw := gateway.New(cfg, cacheManager, logger)

	// Test GetProxy method
	proxy := gw.GetProxy()
	assert.NotNil(t, proxy)
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

	gw := gateway.New(cfg, cacheManager, logger)

	// Test GetBackendManager method
	backendManager := gw.GetBackendManager()
	assert.NotNil(t, backendManager)
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

	gw := gateway.New(cfg, cacheManager, logger)

	// Test GetCircuitManager method
	circuitManager := gw.GetCircuitManager()
	assert.NotNil(t, circuitManager)
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

			gw := gateway.New(cfg, cacheManager, logger)

			// Verify backends were added (for valid cases)
			if !tt.expectError {
				for _, backendCfg := range tt.backendCfgs {
					backend, err := gw.GetBackendManager().GetBackend(backendCfg.Name)
					require.NoError(t, err)
					assert.NotNil(t, backend)
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

	gw := gateway.New(cfg, cacheManager, logger)

	// Verify that health checks are started
	// We can't directly test the goroutine, but we can verify the backend manager exists
	assert.NotNil(t, gw.GetBackendManager())

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

	gateway := gateway.New(cfg, cacheManager, logger)

	// Should still create gateway successfully
	assert.NotNil(t, gateway)
	assert.NotNil(t, gateway.GetBackendManager())
	assert.NotNil(t, gateway.GetCircuitManager())
	assert.NotNil(t, gateway.GetProxy())
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
	gw := gateway.New(cfg, nil, &logger.Logger{})
	assert.NotNil(t, gw)
	assert.Nil(t, gw.GetCacheManager())

	// Test with nil logger
	gw2 := gateway.New(cfg, NewMockCache(), nil)
	assert.NotNil(t, gw2)
	assert.Nil(t, gw2.GetLogger())
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

	gateway := gateway.New(cfg, cacheManager, logger)

	// Test concurrent access to gateway components
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			// Access different components concurrently
			assert.NotNil(t, gateway.GetBackendManager())
			assert.NotNil(t, gateway.GetCircuitManager())
			assert.NotNil(t, gateway.GetProxy())
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
