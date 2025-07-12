package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"kalshi/internal/api/routes"
	"kalshi/internal/auth"
	"kalshi/internal/cache"
	"kalshi/internal/config"
	"kalshi/internal/gateway"
	"kalshi/internal/ratelimit"
	"kalshi/internal/storage"
	"kalshi/pkg/logger"

	"github.com/gin-gonic/gin"
)

// TestConfig holds test configuration and components
type TestConfig struct {
	Config        *config.Config
	Gateway       *gateway.Gateway
	Storage       storage.Storage
	CacheManager  *cache.Manager
	JWTManager    *auth.JWTManager
	APIKeyManager *auth.APIKeyManager
	Limiter       *ratelimit.Limiter
	Logger        *logger.Logger
	Router        http.Handler
	TestServer    *httptest.Server
}

// TestData represents the structure of users.json
type TestData struct {
	Users   []UserInfo            `json:"users"`
	APIKeys map[string]APIKeyData `json:"api_keys"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type APIKeyData struct {
	UserID    string `json:"user_id"`
	RateLimit int    `json:"rate_limit"`
	Enabled   bool   `json:"enabled"`
}

// loadTestData loads API keys from the test data file
func loadTestData(t *testing.T, storage storage.Storage) error {
	// Read test data file
	dataPath := filepath.Join("testdata", "users.json")
	// Loading test data from file
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return fmt.Errorf("failed to read test data file: %w", err)
	}

	var testData TestData
	if err := json.Unmarshal(data, &testData); err != nil {
		return fmt.Errorf("failed to parse test data: %w", err)
	}

	ctx := context.Background()

	// Load API keys into storage
	for apiKey, keyData := range testData.APIKeys {
		// Create API key info
		keyInfo := auth.APIKeyInfo{
			UserID:    keyData.UserID,
			RateLimit: keyData.RateLimit,
			Enabled:   keyData.Enabled,
			CreatedAt: time.Now(),
		}

		// Serialize to JSON
		keyDataBytes, err := json.Marshal(keyInfo)
		if err != nil {
			return fmt.Errorf("failed to serialize API key data: %w", err)
		}

		// Store in storage
		storageKey := "apikey:" + apiKey
		// Setting API key in storage
		if err := storage.Set(ctx, storageKey, string(keyDataBytes), 24*time.Hour); err != nil {
			return fmt.Errorf("failed to store API key %s: %w", apiKey, err)
		}

		// Add to user's key list
		userKey := "userkeys:" + keyData.UserID
		var keyList []string

		// Try to get existing list
		existingData, err := storage.Get(ctx, userKey)
		if err == nil {
			if err := json.Unmarshal([]byte(existingData), &keyList); err != nil {
				return fmt.Errorf("failed to parse existing user key list: %w", err)
			}
		}

		// Add new key if not already present
		found := false
		for _, existingKey := range keyList {
			if existingKey == apiKey {
				found = true
				break
			}
		}
		if !found {
			keyList = append(keyList, apiKey)
		}

		// Store updated list
		listDataBytes, err := json.Marshal(keyList)
		if err != nil {
			return fmt.Errorf("failed to serialize user key list: %w", err)
		}

		if err := storage.Set(ctx, userKey, string(listDataBytes), 24*time.Hour); err != nil {
			return fmt.Errorf("failed to store user key list: %w", err)
		}
	}

	return nil
}

// SetupTestEnvironment creates a complete test environment
func SetupTestEnvironment(t *testing.T) *TestConfig {
	t.Helper()

	// Ensure Gin is in test mode to avoid global state issues
	gin.SetMode(gin.TestMode)

	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         8081,
			Host:         "localhost",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  10 * time.Second,
		},
		Auth: config.AuthConfig{
			JWT: config.JWTConfig{
				Secret:        "test-secret-key-for-integration-tests-32-chars",
				AccessExpiry:  time.Hour,
				RefreshExpiry: 24 * time.Hour,
			},
			APIKey: config.APIKeyConfig{
				Enabled: true,
				Header:  "X-API-Key",
			},
		},
		RateLimit: config.RateLimitConfig{
			DefaultRate:     1000,
			BurstCapacity:   100,
			Storage:         "memory",
			CleanupInterval: 30 * time.Second,
		},
		Cache: config.CacheConfig{
			Memory: config.MemoryConfig{
				MaxSize: 100,
				TTL:     time.Minute,
			},
		},
		Circuit: config.CircuitConfig{
			FailureThreshold: 3,
			RecoveryTimeout:  10 * time.Second,
			MaxRequests:      2,
		},
		Logging: config.LoggingConfig{
			Level:  "error",
			Format: "console",
		},
		Metrics: config.MetricsConfig{
			Enabled: true, // Enable for tests
		},
	}

	// Initialize logger
	log, err := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize storage
	stor := storage.NewMemoryStorage()

	// Load test data (API keys)
	if err := loadTestData(t, stor); err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	// Initialize cache
	l1Cache := cache.NewMemoryCache(cfg.Cache.Memory.MaxSize, cfg.Cache.Memory.TTL)
	cacheManager := cache.NewManager(l1Cache, nil, true)

	// Initialize authentication
	jwtManager := auth.NewJWTManager(cfg.Auth.JWT.Secret, cfg.Auth.JWT.AccessExpiry, cfg.Auth.JWT.RefreshExpiry)
	apiKeyManager := auth.NewAPIKeyManager(stor)

	// Initialize rate limiter
	limiter := ratelimit.NewLimiter(stor, cfg.RateLimit.DefaultRate, cfg.RateLimit.BurstCapacity)

	// Create mock backend server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleMockBackend(w, r)
	}))

	// Update config with test server URL
	cfg.Backend = []config.BackendConfig{
		{
			Name:        "test-backend",
			URL:         testServer.URL,
			HealthCheck: "/health",
			Weight:      100,
		},
		{
			Name:        "slow-backend",
			URL:         testServer.URL + "/slow",
			HealthCheck: "/health",
			Weight:      50,
		},
		{
			Name:        "failing-backend",
			URL:         testServer.URL + "/fail",
			HealthCheck: "/health",
			Weight:      25,
		},
	}

	cfg.Routes = []config.RouteConfig{
		{
			Path:      "/api/v1/users/*",
			Backend:   "test-backend",
			Methods:   []string{"GET", "POST", "PUT", "DELETE"},
			RateLimit: 100,
			CacheTTL:  30 * time.Second,
		},
		{
			Path:      "/api/v1/slow/*",
			Backend:   "slow-backend",
			Methods:   []string{"GET"},
			RateLimit: 50,
			CacheTTL:  60 * time.Second,
		},
		{
			Path:      "/api/v1/fail/*",
			Backend:   "failing-backend",
			Methods:   []string{"GET"},
			RateLimit: 25,
			CacheTTL:  0, // No caching for failing backend
		},
	}

	// Initialize gateway
	gw := gateway.New(cfg, cacheManager, log)

	// Setup router
	routerConfig := &routes.RouterConfig{
		Config:        cfg,
		Gateway:       gw,
		Limiter:       limiter,
		JWTManager:    jwtManager,
		APIKeyManager: apiKeyManager,
		Logger:        log,
	}

	router := routes.SetupTestRouter(routerConfig)

	return &TestConfig{
		Config:        cfg,
		Gateway:       gw,
		Storage:       stor,
		CacheManager:  cacheManager,
		JWTManager:    jwtManager,
		APIKeyManager: apiKeyManager,
		Limiter:       limiter,
		Logger:        log,
		Router:        router,
		TestServer:    testServer,
	}
}

// CleanupTestEnvironment cleans up test resources
func CleanupTestEnvironment(tc *TestConfig) {
	if tc.TestServer != nil {
		tc.TestServer.Close()
	}
	if tc.Storage != nil {
		tc.Storage.Close()
	}
}

// HandleMockBackend handles requests to the mock backend
func HandleMockBackend(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Simulate different backend behaviors
	switch {
	case r.URL.Path == "/health":
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "healthy", "service": "mock-backend"}`)

	case r.URL.Path == "/slow/health":
		time.Sleep(2 * time.Second) // Simulate slow response
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "healthy", "service": "slow-backend"}`)

	case r.URL.Path == "/fail/health":
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "backend unavailable"}`)

	case r.URL.Path == "/slow":
		time.Sleep(1 * time.Second) // Simulate slow response
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "slow response", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))

	case r.URL.Path == "/fail":
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "simulated failure"}`)

	default:
		// Default successful response
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"method":    r.Method,
			"path":      r.URL.Path,
			"query":     r.URL.RawQuery,
			"headers":   r.Header,
			"timestamp": time.Now().Format(time.RFC3339),
		}

		// Echo request body for POST/PUT requests
		if r.Method == "POST" || r.Method == "PUT" {
			// Simple echo for testing
			response["body_received"] = true
		}

		fmt.Fprintf(w, `{"data": %v}`, response)
	}
}

// GenerateTestJWT generates a JWT token for testing
func GenerateTestJWT(jwtManager *auth.JWTManager, userID, role string) (string, error) {
	tokenPair, err := jwtManager.GenerateTokenPair(userID, role)
	if err != nil {
		return "", err
	}
	return tokenPair.AccessToken, nil
}

// WaitForCondition waits for a condition to be true or timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for condition: %s", message)
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}
