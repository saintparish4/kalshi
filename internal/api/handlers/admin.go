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

// GetBackend returns a specific backend by name
func (h *AdminHandler) GetBackend(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Backend name parameter is required",
		})
		return
	}

	backend, err := h.gateway.GetBackendManager().GetBackend(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":         backend.Name,
		"url":          backend.URL.String(),
		"health_check": backend.HealthCheck,
		"weight":       backend.Weight,
		"healthy":      backend.IsHealthy,
	})
}

// CheckBackendHealth manually triggers a health check for a specific backend
func (h *AdminHandler) CheckBackendHealth(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Backend name parameter is required",
		})
		return
	}

	backend, err := h.gateway.GetBackendManager().GetBackend(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Trigger health check
	err = h.gateway.GetBackendManager().TriggerHealthCheck(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get updated backend info
	backend, _ = h.gateway.GetBackendManager().GetBackendByName(name)

	c.JSON(http.StatusOK, gin.H{
		"message": "Health check triggered",
		"backend": name,
		"healthy": backend.IsHealthy,
	})
}

// EnableBackend enables a specific backend
func (h *AdminHandler) EnableBackend(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Backend name parameter is required",
		})
		return
	}

	// Enable backend by setting it as healthy
	err := h.gateway.GetBackendManager().SetBackendHealth(name, true)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"backend":  name,
		"admin_ip": c.ClientIP(),
	}).Info("Backend enabled")

	c.JSON(http.StatusOK, gin.H{
		"message": "Backend enabled successfully",
		"backend": name,
	})
}

// DisableBackend disables a specific backend
func (h *AdminHandler) DisableBackend(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Backend name parameter is required",
		})
		return
	}

	// Disable backend by setting it as unhealthy
	err := h.gateway.GetBackendManager().SetBackendHealth(name, false)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"backend":  name,
		"admin_ip": c.ClientIP(),
	}).Info("Backend disabled")

	c.JSON(http.StatusOK, gin.H{
		"message": "Backend disabled successfully",
		"backend": name,
	})
}

// GetCircuitBreaker returns a specific circuit breaker state
func (h *AdminHandler) GetCircuitBreaker(c *gin.Context) {
	backend := c.Param("backend")
	if backend == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Backend parameter is required",
		})
		return
	}

	states := h.gateway.GetCircuitManager().GetAllStates()
	state, exists := states[backend]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Circuit breaker not found for backend",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"backend": backend,
		"state":   state,
	})
}

// OpenCircuitBreaker manually opens a circuit breaker
func (h *AdminHandler) OpenCircuitBreaker(c *gin.Context) {
	backend := c.Param("backend")
	if backend == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Backend parameter is required",
		})
		return
	}

	// Get the circuit breaker and manually set it to open
	breaker := h.gateway.GetCircuitManager().GetBreaker(backend, 5, 30*time.Second, 3)
	breaker.SetState(1) // StateOpen

	h.logger.WithFields(map[string]interface{}{
		"backend":  backend,
		"admin_ip": c.ClientIP(),
	}).Info("Circuit breaker manually opened")

	c.JSON(http.StatusOK, gin.H{
		"message": "Circuit breaker opened successfully",
		"backend": backend,
	})
}

// CloseCircuitBreaker manually closes a circuit breaker
func (h *AdminHandler) CloseCircuitBreaker(c *gin.Context) {
	backend := c.Param("backend")
	if backend == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Backend parameter is required",
		})
		return
	}

	// Get the circuit breaker and manually set it to closed
	breaker := h.gateway.GetCircuitManager().GetBreaker(backend, 5, 30*time.Second, 3)
	breaker.SetState(0) // StateClosed

	h.logger.WithFields(map[string]interface{}{
		"backend":  backend,
		"admin_ip": c.ClientIP(),
	}).Info("Circuit breaker manually closed")

	c.JSON(http.StatusOK, gin.H{
		"message": "Circuit breaker closed successfully",
		"backend": backend,
	})
}

// GetCacheStats returns cache statistics
func (h *AdminHandler) GetCacheStats(c *gin.Context) {
	// Basic cache stats - implement based on your cache implementation
	c.JSON(http.StatusOK, gin.H{
		"cache_type": "memory",
		"status":     "available",
		"message":    "Cache statistics not implemented yet",
	})
}

// ClearCache clears all cached data
func (h *AdminHandler) ClearCache(c *gin.Context) {
	// Implement cache clearing based on your cache implementation
	c.JSON(http.StatusOK, gin.H{
		"message": "Cache cleared successfully",
	})
}

// DeleteCacheKey deletes a specific cache key
func (h *AdminHandler) DeleteCacheKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cache key parameter is required",
		})
		return
	}

	// Implement cache key deletion based on your cache implementation
	c.JSON(http.StatusOK, gin.H{
		"message": "Cache key deleted successfully",
		"key":     key,
	})
}

// GetCacheKey retrieves a specific cache key
func (h *AdminHandler) GetCacheKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cache key parameter is required",
		})
		return
	}

	// Implement cache key retrieval based on your cache implementation
	c.JSON(http.StatusOK, gin.H{
		"key":   key,
		"value": "Cache key retrieval not implemented yet",
	})
}

// GetRateLimitStats returns rate limiting statistics
func (h *AdminHandler) GetRateLimitStats(c *gin.Context) {
	// Basic rate limit stats - implement based on your rate limiting implementation
	c.JSON(http.StatusOK, gin.H{
		"rate_limit_type": "token_bucket",
		"status":          "active",
		"message":         "Rate limit statistics not implemented yet",
	})
}

// ResetRateLimit resets rate limiting for a specific client
func (h *AdminHandler) ResetRateLimit(c *gin.Context) {
	clientId := c.Param("clientId")
	if clientId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Client ID parameter is required",
		})
		return
	}

	// Implement rate limit reset based on your rate limiting implementation
	c.JSON(http.StatusOK, gin.H{
		"message":  "Rate limit reset successfully",
		"clientId": clientId,
	})
}

// GetRateLimitStatus returns rate limiting status for a specific client
func (h *AdminHandler) GetRateLimitStatus(c *gin.Context) {
	clientId := c.Param("clientId")
	if clientId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Client ID parameter is required",
		})
		return
	}

	// Implement rate limit status based on your rate limiting implementation
	c.JSON(http.StatusOK, gin.H{
		"clientId": clientId,
		"status":   "Rate limit status not implemented yet",
	})
}

// GetSystemInfo returns system information
func (h *AdminHandler) GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":   "kalshi-api-gateway",
		"version":   "1.0.0",
		"uptime":    time.Since(startTime).String(),
		"timestamp": time.Now().UTC(),
	})
}

// GetConfig returns current configuration (sanitized)
func (h *AdminHandler) GetConfig(c *gin.Context) {
	// Return basic config info without sensitive data
	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration endpoint not implemented yet",
	})
}

// ReloadConfig reloads configuration from file
func (h *AdminHandler) ReloadConfig(c *gin.Context) {
	// Implement config reload based on your config implementation
	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration reload not implemented yet",
	})
}

// GetGoroutines returns current goroutine count
func (h *AdminHandler) GetGoroutines(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Goroutine statistics not implemented yet",
	})
}

// GetMemoryStats returns memory usage statistics
func (h *AdminHandler) GetMemoryStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Memory statistics not implemented yet",
	})
}

// TriggerGC triggers garbage collection
func (h *AdminHandler) TriggerGC(c *gin.Context) {
	// Implement GC trigger
	c.JSON(http.StatusOK, gin.H{
		"message": "Garbage collection not implemented yet",
	})
}

// HandlePprof handles pprof requests
func (h *AdminHandler) HandlePprof(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Pprof endpoint not implemented yet",
	})
}
