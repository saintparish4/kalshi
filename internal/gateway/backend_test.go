package gateway

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBackendManager(t *testing.T) {
	bm := NewBackendManager()
	assert.NotNil(t, bm)
	assert.NotNil(t, bm.backends)
	assert.Empty(t, bm.backends)
}

func TestBackendManager_AddBackend(t *testing.T) {
	bm := NewBackendManager()

	tests := []struct {
		name        string
		backendName string
		urlStr      string
		healthCheck string
		weight      int
		expectError bool
	}{
		{
			name:        "valid backend",
			backendName: "test-backend",
			urlStr:      "http://localhost:8080",
			healthCheck: "/health",
			weight:      1,
			expectError: false,
		},
		{
			name:        "invalid URL",
			backendName: "invalid-backend",
			urlStr:      "://invalid-url",
			healthCheck: "/health",
			weight:      1,
			expectError: true,
		},
		{
			name:        "empty URL",
			backendName: "empty-backend",
			urlStr:      "://",
			healthCheck: "/health",
			weight:      1,
			expectError: true,
		},
		{
			name:        "HTTPS URL",
			backendName: "https-backend",
			urlStr:      "https://api.example.com",
			healthCheck: "/health",
			weight:      2,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bm.AddBackend(tt.backendName, tt.urlStr, tt.healthCheck, tt.weight)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify backend was added
				backend, exists := bm.backends[tt.backendName]
				assert.True(t, exists)
				assert.Equal(t, tt.backendName, backend.Name)
				assert.Equal(t, tt.weight, backend.Weight)
				assert.Equal(t, tt.healthCheck, backend.HealthCheck)
				assert.True(t, backend.IsHealthy)
			}
		})
	}
}

func TestBackendManager_GetBackend(t *testing.T) {
	bm := NewBackendManager()

	// Add a test backend
	err := bm.AddBackend("test-backend", "http://localhost:8080", "/health", 1)
	require.NoError(t, err)

	tests := []struct {
		name        string
		backendName string
		expectError bool
	}{
		{
			name:        "existing backend",
			backendName: "test-backend",
			expectError: false,
		},
		{
			name:        "non-existing backend",
			backendName: "non-existing",
			expectError: true,
		},
		{
			name:        "empty backend name",
			backendName: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := bm.GetBackend(tt.backendName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, backend)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, backend)
				assert.Equal(t, tt.backendName, backend.Name)
			}
		})
	}
}

func TestBackendManager_GetBackend_Unhealthy(t *testing.T) {
	bm := NewBackendManager()

	// Add a test backend
	err := bm.AddBackend("test-backend", "http://localhost:8080", "/health", 1)
	require.NoError(t, err)

	// Mark backend as unhealthy
	backend, exists := bm.backends["test-backend"]
	require.True(t, exists)
	backend.mu.Lock()
	backend.IsHealthy = false
	backend.mu.Unlock()

	// Try to get the unhealthy backend
	backend, err = bm.GetBackend("test-backend")
	assert.Error(t, err)
	assert.Nil(t, backend)
	assert.Contains(t, err.Error(), "unhealthy")
}

func TestBackendManager_StartHealthChecks(t *testing.T) {
	// Create a test server that responds with 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	bm := NewBackendManager()

	// Add a test backend with the test server URL
	err := bm.AddBackend("test-backend", server.URL, "/health", 1)
	require.NoError(t, err)

	// Start health checks with a short interval
	bm.StartHealthChecks(100 * time.Millisecond)

	// Give some time for health checks to run
	time.Sleep(300 * time.Millisecond)

	// Verify that health checks are running (backend should still be healthy)
	backend, err := bm.GetBackend("test-backend")
	assert.NoError(t, err)
	assert.NotNil(t, backend)
}

func TestBackendManager_performHealthChecks(t *testing.T) {
	// Create test servers that respond with 200 OK
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server2.Close()

	bm := NewBackendManager()

	// Add test backends with the test server URLs
	err := bm.AddBackend("backend1", server1.URL, "/health", 1)
	require.NoError(t, err)
	err = bm.AddBackend("backend2", server2.URL, "/health", 2)
	require.NoError(t, err)

	// Perform health checks
	bm.performHealthChecks()

	// Give some time for health checks to complete
	time.Sleep(100 * time.Millisecond)

	// Verify backends are still accessible
	backend1, err := bm.GetBackend("backend1")
	assert.NoError(t, err)
	assert.NotNil(t, backend1)

	backend2, err := bm.GetBackend("backend2")
	assert.NoError(t, err)
	assert.NotNil(t, backend2)
}

func TestBackendManager_checkBackendHealth(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	bm := NewBackendManager()

	// Add backend with the test server URL
	err := bm.AddBackend("test-backend", server.URL, "/health", 1)
	require.NoError(t, err)

	// Get the backend
	backend, exists := bm.backends["test-backend"]
	require.True(t, exists)

	// Check health
	bm.checkBackendHealth(backend)

	// Give some time for the health check to complete
	time.Sleep(50 * time.Millisecond)

	// Verify backend is still healthy
	backend.mu.RLock()
	isHealthy := backend.IsHealthy
	backend.mu.RUnlock()
	assert.True(t, isHealthy)
}

func TestBackendManager_checkBackendHealth_Unhealthy(t *testing.T) {
	// Create a test server that returns 500 error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	bm := NewBackendManager()

	// Add backend with the test server URL
	err := bm.AddBackend("test-backend", server.URL, "/health", 1)
	require.NoError(t, err)

	// Get the backend
	backend, exists := bm.backends["test-backend"]
	require.True(t, exists)

	// Check health
	bm.checkBackendHealth(backend)

	// Give some time for the health check to complete
	time.Sleep(50 * time.Millisecond)

	// Verify backend is marked as unhealthy
	backend.mu.RLock()
	isHealthy := backend.IsHealthy
	backend.mu.RUnlock()
	assert.False(t, isHealthy)
}

func TestBackendManager_checkBackendHealth_NoHealthCheck(t *testing.T) {
	bm := NewBackendManager()

	// Add backend without health check
	err := bm.AddBackend("test-backend", "http://localhost:8080", "", 1)
	require.NoError(t, err)

	// Get the backend
	backend, exists := bm.backends["test-backend"]
	require.True(t, exists)

	// Check health (should not panic and should not change health status)
	bm.checkBackendHealth(backend)

	// Verify backend is still healthy (default state)
	backend.mu.RLock()
	isHealthy := backend.IsHealthy
	backend.mu.RUnlock()
	assert.True(t, isHealthy)
}

func TestBackendManager_updateBackendHealth(t *testing.T) {
	bm := NewBackendManager()

	// Add a test backend
	err := bm.AddBackend("test-backend", "http://localhost:8080", "/health", 1)
	require.NoError(t, err)

	backend, exists := bm.backends["test-backend"]
	require.True(t, exists)

	// Test updating to unhealthy
	bm.updateBackendHealth(backend, false)
	backend.mu.RLock()
	assert.False(t, backend.IsHealthy)
	backend.mu.RUnlock()

	// Test updating to healthy
	bm.updateBackendHealth(backend, true)
	backend.mu.RLock()
	assert.True(t, backend.IsHealthy)
	backend.mu.RUnlock()
}

func TestBackend_ConcurrentAccess(t *testing.T) {
	bm := NewBackendManager()

	// Add a test backend
	err := bm.AddBackend("test-backend", "http://localhost:8080", "/health", 1)
	require.NoError(t, err)

	backend, exists := bm.backends["test-backend"]
	require.True(t, exists)

	// Test concurrent access to backend health status
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			backend.mu.RLock()
			_ = backend.IsHealthy
			backend.mu.RUnlock()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestBackendManager_ConcurrentAddGet(t *testing.T) {
	bm := NewBackendManager()

	// Test concurrent add and get operations
	done := make(chan bool, 10)
	for i := 0; i < 5; i++ {
		go func(id int) {
			backendName := fmt.Sprintf("backend-%d", id)
			err := bm.AddBackend(backendName, "http://localhost:8080", "/health", id)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	for i := 5; i < 10; i++ {
		go func(id int) {
			backendName := fmt.Sprintf("backend-%d", id-5)
			backend, err := bm.GetBackend(backendName)
			// This might fail if the backend hasn't been added yet
			if err == nil {
				assert.NotNil(t, backend)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestBackendManager_HealthCheckTimeout(t *testing.T) {
	// Create a test server that takes too long to respond
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(6 * time.Second) // Longer than the 5-second timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	bm := NewBackendManager()

	// Add backend with the slow test server URL
	serverURL := server.URL
	err := bm.AddBackend("test-backend", serverURL, "/health", 1)
	require.NoError(t, err)

	// Get the backend
	backend, exists := bm.backends["test-backend"]
	require.True(t, exists)

	// Check health (should timeout and mark as unhealthy)
	bm.checkBackendHealth(backend)

	// Give some time for the health check to complete
	time.Sleep(100 * time.Millisecond)

	// Verify backend is marked as unhealthy due to timeout
	backend.mu.RLock()
	isHealthy := backend.IsHealthy
	backend.mu.RUnlock()
	assert.False(t, isHealthy)
}

func TestBackendManager_InvalidURL(t *testing.T) {
	bm := NewBackendManager()

	// Try to add backend with invalid URL
	err := bm.AddBackend("test-backend", "://invalid-url", "/health", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse backend URL")

	// Verify backend was not added
	_, exists := bm.backends["test-backend"]
	assert.False(t, exists)
}

func TestBackendManager_EmptyBackends(t *testing.T) {
	bm := NewBackendManager()

	// Test with no backends
	backend, err := bm.GetBackend("non-existing")
	assert.Error(t, err)
	assert.Nil(t, backend)
	assert.Contains(t, err.Error(), "not found")

	// Perform health checks on empty backend list
	bm.performHealthChecks()
	// Should not panic
}

func TestBackend_ThreadSafety(t *testing.T) {
	backend := &Backend{
		Name:        "test-backend",
		HealthCheck: "/health",
		Weight:      1,
		IsHealthy:   true,
	}

	// Test concurrent read/write access
	done := make(chan bool, 20)
	for i := 0; i < 10; i++ {
		go func() {
			backend.mu.RLock()
			_ = backend.IsHealthy
			backend.mu.RUnlock()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			backend.mu.Lock()
			backend.IsHealthy = !backend.IsHealthy
			backend.mu.Unlock()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}
}
