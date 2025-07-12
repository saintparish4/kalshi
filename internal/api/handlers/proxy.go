package handlers

import (
	"net/http"
	"strings"
	"time"

	"kalshi/internal/config"
	"kalshi/internal/gateway"
	"kalshi/pkg/logger"

	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	gateway *gateway.Gateway
	config  *config.Config
	logger  *logger.Logger
	proxy   interface {
		ServeHTTP(http.ResponseWriter, *http.Request, string, time.Duration)
	}
}

func NewProxyHandler(gateway *gateway.Gateway, config *config.Config, logger *logger.Logger) *ProxyHandler {
	return &ProxyHandler{
		gateway: gateway,
		config:  config,
		logger:  logger,
		proxy:   nil,
	}
}

// HandleRequest processess all proxied requests
func (h *ProxyHandler) HandleRequest(c *gin.Context) {
	path := c.Request.URL.Path
	method := c.Request.Method

	// Find matching route
	route := h.findMatchingRoute(path, method)
	if route == nil {
		h.logger.WithFields(map[string]interface{}{
			"path":   path,
			"method": method,
		}).Warn("No matching route found")

		c.JSON(http.StatusNotFound, gin.H{
			"error": "Route not found",
			"path":  path,
		})
		return
	}

	// Validate HTTP method
	if !h.isMethodAllowed(route.Methods, method) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error":           "Method not allowed",
			"allowed_methods": route.Methods,
		})
		return
	}

	// Set Backend in context for metrics
	c.Set("backend", route.Backend)
	c.Set("route_path", route.Path)

	// Get cache TTL (use route-specific or default)
	cacheTTL := route.CacheTTL
	if cacheTTL == 0 {
		cacheTTL = 5 * time.Minute // default cache TTL
	}

	// Log the request
	h.logger.WithFields(map[string]interface{}{
		"path":      path,
		"method":    method,
		"backend":   route.Backend,
		"cache_ttl": cacheTTL.String(),
		"client_ip": c.ClientIP(),
	}).Info("Proxying request")

	// Proxy the request
	if h.proxy != nil {
		h.proxy.ServeHTTP(c.Writer, c.Request, route.Backend, cacheTTL)
	} else {
		h.gateway.GetProxy().ServeHTTP(c.Writer, c.Request, route.Backend, cacheTTL)
	}
}

// findMatchingRoute finds the first route that matches the given path
func (h *ProxyHandler) findMatchingRoute(path, method string) *config.RouteConfig {
	for _, route := range h.config.Routes {
		if h.matchRoute(path, route.Path) {
			return &route
		}
	}
	return nil
}

// matchRoute performs path matching with wildcard support
func (h *ProxyHandler) matchRoute(path, pattern string) bool {
	// Exact match
	if path == pattern {
		return true
	}

	// Wildcard matching
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	// Path parameter matching (basic implementation)
	if strings.HasPrefix(pattern, ":") {
		return h.matchWithParams(path, pattern)
	}
	return false
}

// matchWithParams handles path parameters like /api/users/:id
func (h *ProxyHandler) matchWithParams(path, pattern string) bool {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")

	if len(pathParts) != len(patternParts) {
		return false
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			// This is a parameter, skip validation
			continue
		}

		if part != pathParts[i] {
			return false
		}
	}
	return true
}

// isMethodAllowed checks if the HTTP method is allowed for the route
func (h *ProxyHandler) isMethodAllowed(allowedMethods []string, method string) bool {
	for _, allowed := range allowedMethods {
		if method == allowed {
			return true
		}
	}
	return false
}

// GetRoutes returns all configured routes (for debugging)
func (h *ProxyHandler) GetRoutes(c *gin.Context) {
	routes := make([]gin.H, 0, len(h.config.Routes))

	for _, route := range h.config.Routes {
		routes = append(routes, gin.H{
			"path":       route.Path,
			"backend":    route.Backend,
			"methods":    route.Methods,
			"rate_limit": route.RateLimit,
			"cache_ttl":  route.CacheTTL.String(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"routes": routes,
		"total":  len(routes),
	})
}

// GetRoute returns a specific route by ID
func (h *ProxyHandler) GetRoute(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Route ID parameter is required",
		})
		return
	}

	// Find route by ID (using path as ID for now)
	for _, route := range h.config.Routes {
		if route.Path == id {
			c.JSON(http.StatusOK, gin.H{
				"path":       route.Path,
				"backend":    route.Backend,
				"methods":    route.Methods,
				"rate_limit": route.RateLimit,
				"cache_ttl":  route.CacheTTL.String(),
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "Route not found",
		"id":    id,
	})
}

// ReloadRoutes reloads route configuration
func (h *ProxyHandler) ReloadRoutes(c *gin.Context) {
	// For now, just return success since we don't have dynamic config reload
	c.JSON(http.StatusOK, gin.H{
		"message": "Route reload not implemented yet",
	})
}
