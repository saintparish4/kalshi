package testing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kalshi/internal/gateway"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBackendManager(t *testing.T) {
	bm := gateway.NewBackendManager()
	assert.NotNil(t, bm)

	// Test that we can get backends (even if empty)
	backends := bm.GetBackends()
	assert.Empty(t, backends)
}

func TestBackendManager_AddBackend(t *testing.T) {
	bm := gateway.NewBackendManager()

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
				// Verify backend was added by trying to get it
				backend, err := bm.GetBackend(tt.backendName)
				assert.NoError(t, err)
				assert.NotNil(t, backend)
			}
		})
	}
}

func TestBackendManager_GetBackend(t *testing.T) {
	bm := gateway.NewBackendManager()

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
			}
		})
	}
}

func TestBackendManager_GetBackend_Unhealthy(t *testing.T) {
	bm := gateway.NewBackendManager()

	// Add a test backend
	err := bm.AddBackend("test-backend", "http://localhost:8080", "/health", 1)
	require.NoError(t, err)

	// Mark backend as unhealthy using the exported method
	err = bm.SetBackendHealth("test-backend", false)
	require.NoError(t, err)

	// Try to get the unhealthy backend
	backend, err := bm.GetBackend("test-backend")
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

	bm := gateway.NewBackendManager()

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

	bm := gateway.NewBackendManager()

	// Add test backends with the test server URLs
	err := bm.AddBackend("backend1", server1.URL, "/health", 1)
	require.NoError(t, err)
	err = bm.AddBackend("backend2", server2.URL, "/health", 2)
	require.NoError(t, err)

	// Trigger health checks manually
	err = bm.TriggerHealthCheck("backend1")
	assert.NoError(t, err)
	err = bm.TriggerHealthCheck("backend2")
	assert.NoError(t, err)

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

	bm := gateway.NewBackendManager()

	// Add backend with the test server URL
	err := bm.AddBackend("test-backend", server.URL, "/health", 1)
	require.NoError(t, err)

	// Trigger health check manually
	err = bm.TriggerHealthCheck("test-backend")
	assert.NoError(t, err)

	// Give some time for the health check to complete
	time.Sleep(50 * time.Millisecond)

	// Verify backend is still healthy
	backend, err := bm.GetBackend("test-backend")
	assert.NoError(t, err)
	assert.NotNil(t, backend)
}

func TestBackendManager_checkBackendHealth_Unhealthy(t *testing.T) {
	// Create a test server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	bm := gateway.NewBackendManager()

	// Add backend with the test server URL
	err := bm.AddBackend("test-backend", server.URL, "/health", 1)
	require.NoError(t, err)

	// Trigger health check manually
	err = bm.TriggerHealthCheck("test-backend")
	assert.NoError(t, err)

	// Give some time for the health check to complete
	time.Sleep(100 * time.Millisecond)

	// Verify backend is now unhealthy
	backend, err := bm.GetBackend("test-backend")
	assert.Error(t, err)
	assert.Nil(t, backend)
}

func TestBackendManager_checkBackendHealth_NoHealthCheck(t *testing.T) {
	bm := gateway.NewBackendManager()

	// Add backend without health check
	err := bm.AddBackend("test-backend", "http://localhost:8080", "", 1)
	require.NoError(t, err)

	// Trigger health check manually (should not fail)
	err = bm.TriggerHealthCheck("test-backend")
	assert.NoError(t, err)

	// Backend should still be accessible
	backend, err := bm.GetBackend("test-backend")
	assert.NoError(t, err)
	assert.NotNil(t, backend)
}

func TestBackendManager_updateBackendHealth(t *testing.T) {
	bm := gateway.NewBackendManager()

	// Add a test backend
	err := bm.AddBackend("test-backend", "http://localhost:8080", "/health", 1)
	require.NoError(t, err)

	// Set backend as unhealthy
	err = bm.SetBackendHealth("test-backend", false)
	assert.NoError(t, err)

	// Verify backend is now unhealthy
	backend, err := bm.GetBackend("test-backend")
	assert.Error(t, err)
	assert.Nil(t, backend)

	// Set backend as healthy again
	err = bm.SetBackendHealth("test-backend", true)
	assert.NoError(t, err)

	// Verify backend is now healthy
	backend, err = bm.GetBackend("test-backend")
	assert.NoError(t, err)
	assert.NotNil(t, backend)
}

func TestBackend_ConcurrentAccess(t *testing.T) {
	bm := gateway.NewBackendManager()

	// Add a test backend
	err := bm.AddBackend("test-backend", "http://localhost:8080", "/health", 1)
	require.NoError(t, err)

	// Test concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			backend, err := bm.GetBackend("test-backend")
			assert.NoError(t, err)
			assert.NotNil(t, backend)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestBackendManager_ConcurrentAddGet(t *testing.T) {
	bm := gateway.NewBackendManager()

	// Test concurrent add and get operations
	done := make(chan bool, 20)
	for i := 0; i < 10; i++ {
		go func(id int) {
			backendName := fmt.Sprintf("backend-%d", id)
			err := bm.AddBackend(backendName, "http://localhost:8080", "/health", 1)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		go func(id int) {
			backendName := fmt.Sprintf("backend-%d", id)
			backend, err := bm.GetBackend(backendName)
			// This might fail if the backend hasn't been added yet
			if err == nil {
				assert.NotNil(t, backend)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestBackendManager_HealthCheckTimeout(t *testing.T) {
	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			time.Sleep(6 * time.Second) // Longer than the 5-second timeout
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	bm := gateway.NewBackendManager()

	// Add backend with the test server URL
	err := bm.AddBackend("test-backend", server.URL, "/health", 1)
	require.NoError(t, err)

	// Trigger health check manually
	err = bm.TriggerHealthCheck("test-backend")
	assert.NoError(t, err)

	// Give some time for the health check to complete
	time.Sleep(100 * time.Millisecond)

	// Verify backend is now unhealthy due to timeout
	backend, err := bm.GetBackend("test-backend")
	assert.Error(t, err)
	assert.Nil(t, backend)
}

func TestBackendManager_InvalidURL(t *testing.T) {
	bm := gateway.NewBackendManager()

	// Try to add backend with invalid URL
	err := bm.AddBackend("invalid-backend", "://invalid-url", "/health", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse backend URL")
}

func TestBackendManager_EmptyBackends(t *testing.T) {
	bm := gateway.NewBackendManager()

	// Test getting backends when none exist
	backends := bm.GetBackends()
	assert.Empty(t, backends)

	healthyBackends := bm.GetHealthyBackends()
	assert.Empty(t, healthyBackends)
}

func TestBackend_ThreadSafety(t *testing.T) {
	bm := gateway.NewBackendManager()

	// Add a test backend
	err := bm.AddBackend("test-backend", "http://localhost:8080", "/health", 1)
	require.NoError(t, err)

	// Test concurrent health status changes
	done := make(chan bool, 20)
	for i := 0; i < 10; i++ {
		go func() {
			err := bm.SetBackendHealth("test-backend", false)
			assert.NoError(t, err)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			err := bm.SetBackendHealth("test-backend", true)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}
}
