package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	Name        string
	URL         *url.URL
	HealthCheck string
	Weight      int
	IsHealthy   bool
	mu          sync.RWMutex
}

type BackendManager struct {
	backends map[string]*Backend
	mu       sync.RWMutex
}

func NewBackendManager() *BackendManager {
	return &BackendManager{
		backends: make(map[string]*Backend),
	}
}

func (bm *BackendManager) AddBackend(name, urlStr, healthCheck string, weight int) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("failed to parse backend URL: %w", err)
	}

	backend := &Backend{
		Name:        name,
		URL:         parsedURL,
		HealthCheck: healthCheck,
		Weight:      weight,
		IsHealthy:   true,
	}

	bm.mu.Lock()
	bm.backends[name] = backend
	bm.mu.Unlock()

	return nil
}

func (bm *BackendManager) GetBackend(name string) (*Backend, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	backend, exists := bm.backends[name]
	if !exists {
		return nil, fmt.Errorf("backend %s not found", name)
	}

	backend.mu.RLock()
	defer backend.mu.RUnlock()

	if !backend.IsHealthy {
		return nil, fmt.Errorf("backend %s is unhealthy", name)
	}

	return backend, nil
}

// GetBackends returns all backends
func (bm *BackendManager) GetBackends() []*Backend {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	backends := make([]*Backend, 0, len(bm.backends))
	for _, backend := range bm.backends {
		backends = append(backends, backend)
	}
	return backends
}

// GetHealthyBackends returns only healthy backends
func (bm *BackendManager) GetHealthyBackends() []*Backend {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	healthyBackends := make([]*Backend, 0)
	for _, backend := range bm.backends {
		backend.mu.RLock()
		if backend.IsHealthy {
			healthyBackends = append(healthyBackends, backend)
		}
		backend.mu.RUnlock()
	}
	return healthyBackends
}

// GetAllBackends returns all backends (healthy and unhealthy)
func (bm *BackendManager) GetAllBackends() []*Backend {
	return bm.GetBackends()
}

func (bm *BackendManager) StartHealthChecks(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			bm.performHealthChecks()
		}
	}()
}

func (bm *BackendManager) performHealthChecks() {
	bm.mu.RLock()
	backends := make([]*Backend, 0, len(bm.backends))
	for _, backend := range bm.backends {
		backends = append(backends, backend)
	}
	bm.mu.RUnlock()

	for _, backend := range backends {
		go bm.checkBackendHealth(backend)
	}
}

func (bm *BackendManager) checkBackendHealth(backend *Backend) {
	if backend.HealthCheck == "" {
		return
	}

	healthURL := backend.URL.String() + backend.HealthCheck

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		bm.updateBackendHealth(backend, false)
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		bm.updateBackendHealth(backend, false)
		return
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode >= 200 && resp.StatusCode < 300
	bm.updateBackendHealth(backend, healthy)
}

func (bm *BackendManager) updateBackendHealth(backend *Backend, healthy bool) {
	backend.mu.Lock()
	backend.IsHealthy = healthy
	backend.mu.Unlock()
}
