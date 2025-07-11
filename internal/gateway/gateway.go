package gateway

import (
	"time"

	"kalshi/internal/cache"
	"kalshi/internal/circuit"
	"kalshi/internal/config"
	"kalshi/pkg/logger"
)

type Gateway struct {
	config         *config.Config
	backendManager *BackendManager
	cacheManager   cache.Cache
	circuitManager *circuit.Manager
	proxy          *Proxy
	logger         *logger.Logger
}

func New(cfg *config.Config, cacheManager cache.Cache, logger *logger.Logger) *Gateway {
	backendManager := NewBackendManager()
	circuitManager := circuit.NewManager()

	proxy := NewProxy(backendManager, cacheManager, circuitManager, logger)

	gateway := &Gateway{
		config:         cfg,
		backendManager: backendManager,
		cacheManager:   cacheManager,
		circuitManager: circuitManager,
		proxy:          proxy,
		logger:         logger,
	}

	// Initialize backends
	gateway.initializeBackends()

	// Start health checks
	gateway.backendManager.StartHealthChecks(30 * time.Second)

	return gateway
}

func (g *Gateway) initializeBackends() {
	for _, backend := range g.config.Backend {
		err := g.backendManager.AddBackend(
			backend.Name,
			backend.URL,
			backend.HealthCheck,
			backend.Weight,
		)
		if err != nil {
			g.logger.Error("Failed to add backend", "backend", backend.Name, "error", err)
		}
	}
}

func (g *Gateway) GetProxy() *Proxy {
	return g.proxy
}

func (g *Gateway) GetBackendManager() *BackendManager {
	return g.backendManager
}

func (g *Gateway) GetCircuitManager() *circuit.Manager {
	return g.circuitManager
}
