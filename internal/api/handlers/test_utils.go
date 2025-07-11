package handlers

import (
	"net/http"
	"time"

	"kalshi/internal/config"
	"kalshi/internal/gateway"
	"kalshi/pkg/logger"
)

// Create a proper mock gateway that can be used as *gateway.Gateway
func createMockGateway() *gateway.Gateway {
	cfg := &config.Config{
		Backend: []config.BackendConfig{},
	}
	gateway := gateway.New(cfg, nil, nil)
	return gateway
}

// Create a proper mock logger that can be used in tests
func createMockLogger() *logger.Logger {
	log, err := logger.New("debug", "console")
	if err != nil {
		return &logger.Logger{}
	}
	return log
}

// Create a mock gateway with test backends and circuit breakers
func createMockGatewayWithBackends(backends []config.BackendConfig) *gateway.Gateway {
	cfg := &config.Config{
		Backend: backends,
	}
	log := createMockLogger()
	g := gateway.New(cfg, nil, log)

	// Add each backend using AddBackend (already done in gateway.New), but ensure circuit breakers exist
	circuitMgr := g.GetCircuitManager()
	for _, b := range backends {
		circuitMgr.GetBreaker(b.Name, 5, 30*time.Second, 3)
	}
	return g
}

// MockProxy implements ServeHTTP to simulate a backend response
// Use this in tests to avoid panics
// Usage: handler := NewProxyHandler(mockGateway, cfg, mockLogger); handler.proxy = createMockProxy()
type MockProxy struct{}

func (m *MockProxy) ServeHTTP(w http.ResponseWriter, r *http.Request, backendName string, cacheTTL time.Duration) {
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte("Mock backend error"))
}

func createMockProxy() *MockProxy {
	return &MockProxy{}
}
