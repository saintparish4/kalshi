package handlers

import (
	"net/http"
	"time"

	"kalshi/internal/gateway"
	"kalshi/pkg/logger"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	gateway *gateway.Gateway
	logger  *logger.Logger
}

func NewAdminHandler(gateway *gateway.Gateway, logger *logger.Logger) *AdminHandler {
	return &AdminHandler{
		gateway: gateway,
		logger:  logger,
	}
}

// GetBackends returns all backend information
func (h *AdminHandler) GetBackends(c *gin.Context) {
	backends := h.gateway.GetBackendManager().GetBackends()

	backendInfo := make([]gin.H, 0, len(backends))
	for _, backend := range backends {
		backendInfo = append(backendInfo, gin.H{
			"name":         backend.Name,
			"url":          backend.URL.String(),
			"health_check": backend.HealthCheck,
			"weight":       backend.Weight,
			"healthy":      backend.IsHealthy,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"backends": backendInfo,
		"total":    len(backends),
	})
}

// GetCircuitBreakers returns circuit breaker states
func (h *AdminHandler) GetCircuitBreakers(c *gin.Context) {
	states := h.gateway.GetCircuitManager().GetAllStates()

	c.JSON(http.StatusOK, gin.H{
		"circuit_breakers": states,
		"total":            len(states),
	})
}

// ResetCircuitBreaker resets a specific circuit breaker
func (h *AdminHandler) ResetCircuitBreaker(c *gin.Context) {
	backend := c.Param("backend")
	if backend == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Backend parameter is required",
		})
		return
	}

	err := h.gateway.GetCircuitManager().ResetBreaker(backend)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"backend":  backend,
		"admin_ip": c.ClientIP(),
	}).Info("Circuit breaker reset")

	c.JSON(http.StatusOK, gin.H{
		"message": "Circuit breaker reset successfully",
		"backend": backend,
	})
}

// GetStats returns gateway statistics
func (h *AdminHandler) GetStats(c *gin.Context) {
	healthyBackends := h.gateway.GetBackendManager().GetHealthyBackends()
	allBackends := h.gateway.GetBackendManager().GetAllBackends()
	circuitStates := h.gateway.GetCircuitManager().GetAllStates()

	openCircuits := 0
	for _, state := range circuitStates {
		if state == 1 { // StateOpen
			openCircuits++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_backends":     len(allBackends),
		"healthy_backends":   len(healthyBackends),
		"unhealthy_backends": len(allBackends) - len(healthyBackends),
		"total_circuits":     len(circuitStates),
		"open_circuits":      openCircuits,
		"uptime":             time.Since(startTime).String(),
	})
}
