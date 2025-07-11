package handlers

import (
	"net/http"
	"time"

	"kalshi/internal/gateway"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	gateway *gateway.Gateway
}

func NewHealthHandler(gateway *gateway.Gateway) *HealthHandler {
	return &HealthHandler{gateway: gateway}
}

// Health returns basic health status
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"service":   "kalshi-api-gateway",
	})
}

// DetailedHealth returns detailed health information
func (h *HealthHandler) DetailedHealth(c *gin.Context) {
	backends := h.gateway.GetBackendManager().GetHealthyBackends()
	circuitStates := h.gateway.GetCircuitManager().GetAllStates()

	backendStatus := make(map[string]interface{})
	for _, backend := range backends {
		backendStatus[backend.Name] = map[string]interface{}{
			"healthy": backend.IsHealthy,
			"url":     backend.URL.String(),
			"weight":  backend.Weight,
		}
	}

	// Get All backends (healthy and unhealthy)
	allBackends := h.gateway.GetBackendManager().GetAllBackends()
	allBackendStatus := make(map[string]interface{})
	for _, backend := range allBackends {
		allBackendStatus[backend.Name] = map[string]interface{}{
			"healthy": backend.IsHealthy,
			"url":     backend.URL.String(),
			"weight":  backend.Weight,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "healthy",
		"timestamp":        time.Now().UTC(),
		"version":          "1.0.0",
		"service":          "kalshi-api-gateway",
		"healthy_backends": backendStatus,
		"all_backends":     allBackendStatus,
		"circuit_breaker":  circuitStates,
		"uptime":           time.Since(startTime).String(),
	})
}

// Readiness check for Kubernetes
func (h *HealthHandler) Readiness(c *gin.Context) {
	healthyBackends := h.gateway.GetBackendManager().GetHealthyBackends()

	if len(healthyBackends) == 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "no healthy backends available",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "ready",
		"healthy_backends": len(healthyBackends),
	})
}

// Liveness check for Kubernetes
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().UTC(),
	})
}

var startTime = time.Now()
