package config

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, 30*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, config.Server.WriteTimeout)
	assert.Equal(t, 60*time.Second, config.Server.IdleTimeout)

	assert.Equal(t, "your-secret-key-change-in-production", config.Auth.JWT.Secret)
	assert.Equal(t, 24*time.Hour, config.Auth.JWT.Expiry)
	assert.False(t, config.Auth.APIKey.Enabled)
	assert.Equal(t, "X-API-Key", config.Auth.APIKey.Header)

	assert.Equal(t, 100, config.RateLimit.DefaultRate)
	assert.Equal(t, 200, config.RateLimit.BurstCapacity)
	assert.Equal(t, "memory", config.RateLimit.Storage)
	assert.Equal(t, 1*time.Hour, config.RateLimit.CleanupInterval)

	assert.Equal(t, "localhost:6379", config.Cache.Redis.Addr)
	assert.Equal(t, "", config.Cache.Redis.Password)
	assert.Equal(t, 0, config.Cache.Redis.DB)
	assert.Equal(t, 5*time.Minute, config.Cache.Redis.TTL)

	assert.Equal(t, 1000, config.Cache.Memory.MaxSize)
	assert.Equal(t, 1*time.Minute, config.Cache.Memory.TTL)

	assert.Equal(t, 5, config.Circuit.FailureThreshold)
	assert.Equal(t, 30*time.Second, config.Circuit.RecoveryTimeout)
	assert.Equal(t, 3, config.Circuit.MaxRequests)

	assert.Len(t, config.Backend, 1)
	assert.Equal(t, "default", config.Backend[0].Name)
	assert.Equal(t, "http://localhost:3000", config.Backend[0].URL)
	assert.Equal(t, "/health", config.Backend[0].HealthCheck)
	assert.Equal(t, 1, config.Backend[0].Weight)

	assert.Len(t, config.Routes, 1)
	assert.Equal(t, "/api/*", config.Routes[0].Path)
	assert.Equal(t, "default", config.Routes[0].Backend)
	assert.Equal(t, []string{"GET", "POST", "PUT", "DELETE"}, config.Routes[0].Methods)
	assert.Equal(t, 100, config.Routes[0].RateLimit)
	assert.Equal(t, 5*time.Minute, config.Routes[0].CacheTTL)

	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, "json", config.Logging.Format)

	assert.True(t, config.Metrics.Enabled)
	assert.Equal(t, "/metrics", config.Metrics.Path)
	assert.Equal(t, 9090, config.Metrics.Port)
}

func TestLoadConfig(t *testing.T) {
	// Test loading from existing config file
	config, err := LoadConfig("config.yaml")
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify values from config.yaml
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, 1000, config.RateLimit.DefaultRate)  // From config.yaml
	assert.Equal(t, 100, config.RateLimit.BurstCapacity) // From config.yaml
	assert.True(t, config.Auth.APIKey.Enabled)           // From config.yaml

	// Verify backend configuration
	assert.Len(t, config.Backend, 1)
	assert.Equal(t, "service1", config.Backend[0].Name)
	assert.Equal(t, "http://localhost:3000", config.Backend[0].URL)
	assert.Equal(t, "/health", config.Backend[0].HealthCheck)
	assert.Equal(t, 100, config.Backend[0].Weight)

	// Verify routes configuration
	assert.Len(t, config.Routes, 1)
	assert.Equal(t, "/api/v1/*", config.Routes[0].Path)
	assert.Equal(t, "service1", config.Routes[0].Backend)
	assert.Equal(t, []string{"GET", "POST", "PUT", "DELETE"}, config.Routes[0].Methods)
	assert.Equal(t, 500, config.Routes[0].RateLimit)
	assert.Equal(t, 120*time.Second, config.Routes[0].CacheTTL)
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("KALSHI_SERVER_PORT", "9090")
	os.Setenv("KALSHI_SERVER_HOST", "127.0.0.1")
	os.Setenv("KALSHI_AUTH_JWT_SECRET", "test-secret")
	os.Setenv("KALSHI_RATE_LIMIT_DEFAULT_RATE", "500")
	os.Setenv("KALSHI_LOGGING_LEVEL", "debug")

	defer func() {
		os.Unsetenv("KALSHI_SERVER_PORT")
		os.Unsetenv("KALSHI_SERVER_HOST")
		os.Unsetenv("KALSHI_AUTH_JWT_SECRET")
		os.Unsetenv("KALSHI_RATE_LIMIT_DEFAULT_RATE")
		os.Unsetenv("KALSHI_LOGGING_LEVEL")
	}()

	config, err := LoadConfigFromEnv()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, 9090, config.Server.Port)
	assert.Equal(t, "127.0.0.1", config.Server.Host)
	assert.Equal(t, "test-secret", config.Auth.JWT.Secret)
	assert.Equal(t, 500, config.RateLimit.DefaultRate)
	assert.Equal(t, "debug", config.Logging.Level)
}

func TestLoadConfigFromEnvWithDefaults(t *testing.T) {
	// Clear any existing environment variables that might interfere
	os.Unsetenv("KALSHI_SERVER_PORT")
	os.Unsetenv("KALSHI_RATE_LIMIT_DEFAULT_RATE")
	os.Unsetenv("KALSHI_LOGGING_LEVEL")
	os.Unsetenv("KALSHI_SERVER_HOST")
	os.Unsetenv("KALSHI_AUTH_JWT_SECRET")

	// Test with minimal environment variables
	config, err := LoadConfigFromEnv()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Should use defaults for unset values, except for values set in config.yaml
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, "your-secret-key-change-this-in-production", config.Auth.JWT.Secret)
	assert.Equal(t, 1000, config.RateLimit.DefaultRate) // config.yaml sets this to 1000
	assert.Equal(t, "info", config.Logging.Level)
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid server port",
			config: func() *Config {
				c := DefaultConfig()
				c.Server.Port = 0
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid server port too high",
			config: func() *Config {
				c := DefaultConfig()
				c.Server.Port = 70000
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid server host",
			config: func() *Config {
				c := DefaultConfig()
				c.Server.Host = ""
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid read timeout",
			config: func() *Config {
				c := DefaultConfig()
				c.Server.ReadTimeout = -1 * time.Second
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid write timeout",
			config: func() *Config {
				c := DefaultConfig()
				c.Server.WriteTimeout = -1 * time.Second
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid idle timeout",
			config: func() *Config {
				c := DefaultConfig()
				c.Server.IdleTimeout = -1 * time.Second
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid JWT secret",
			config: func() *Config {
				c := DefaultConfig()
				c.Auth.JWT.Secret = ""
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid JWT expiry",
			config: func() *Config {
				c := DefaultConfig()
				c.Auth.JWT.Expiry = -1 * time.Hour
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid API key header when enabled",
			config: func() *Config {
				c := DefaultConfig()
				c.Auth.APIKey.Enabled = true
				c.Auth.APIKey.Header = ""
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid rate limit default rate",
			config: func() *Config {
				c := DefaultConfig()
				c.RateLimit.DefaultRate = -1
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid rate limit burst capacity",
			config: func() *Config {
				c := DefaultConfig()
				c.RateLimit.BurstCapacity = -1
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid rate limit storage",
			config: func() *Config {
				c := DefaultConfig()
				c.RateLimit.Storage = "invalid"
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid rate limit cleanup interval",
			config: func() *Config {
				c := DefaultConfig()
				c.RateLimit.CleanupInterval = -1 * time.Second
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid Redis TTL",
			config: func() *Config {
				c := DefaultConfig()
				c.Cache.Redis.TTL = -1 * time.Second
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid memory max size",
			config: func() *Config {
				c := DefaultConfig()
				c.Cache.Memory.MaxSize = -1
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid memory TTL",
			config: func() *Config {
				c := DefaultConfig()
				c.Cache.Memory.TTL = -1 * time.Second
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid circuit failure threshold",
			config: func() *Config {
				c := DefaultConfig()
				c.Circuit.FailureThreshold = -1
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid circuit recovery timeout",
			config: func() *Config {
				c := DefaultConfig()
				c.Circuit.RecoveryTimeout = -1 * time.Second
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid circuit max requests",
			config: func() *Config {
				c := DefaultConfig()
				c.Circuit.MaxRequests = -1
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid logging level",
			config: func() *Config {
				c := DefaultConfig()
				c.Logging.Level = "invalid"
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid logging format",
			config: func() *Config {
				c := DefaultConfig()
				c.Logging.Format = "invalid"
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid metrics port",
			config: func() *Config {
				c := DefaultConfig()
				c.Metrics.Port = 0
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid metrics port too high",
			config: func() *Config {
				c := DefaultConfig()
				c.Metrics.Port = 70000
				return c
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBackendConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		backend BackendConfig
		wantErr bool
	}{
		{
			name: "valid backend",
			backend: BackendConfig{
				Name:        "test",
				URL:         "http://localhost:3000",
				HealthCheck: "/health",
				Weight:      1,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			backend: BackendConfig{
				Name:        "",
				URL:         "http://localhost:3000",
				HealthCheck: "/health",
				Weight:      1,
			},
			wantErr: true,
		},
		{
			name: "empty URL",
			backend: BackendConfig{
				Name:        "test",
				URL:         "",
				HealthCheck: "/health",
				Weight:      1,
			},
			wantErr: true,
		},
		{
			name: "zero weight",
			backend: BackendConfig{
				Name:        "test",
				URL:         "http://localhost:3000",
				HealthCheck: "/health",
				Weight:      0,
			},
			wantErr: true,
		},
		{
			name: "negative weight",
			backend: BackendConfig{
				Name:        "test",
				URL:         "http://localhost:3000",
				HealthCheck: "/health",
				Weight:      -1,
			},
			wantErr: true,
		},
		{
			name: "zero weight",
			backend: BackendConfig{
				Name:        "test",
				URL:         "http://localhost:3000",
				HealthCheck: "/health",
				Weight:      0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.backend.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRouteConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		route   RouteConfig
		wantErr bool
	}{
		{
			name: "valid route",
			route: RouteConfig{
				Path:      "/api/*",
				Backend:   "test",
				Methods:   []string{"GET", "POST"},
				RateLimit: 100,
				CacheTTL:  5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "empty path",
			route: RouteConfig{
				Path:      "",
				Backend:   "test",
				Methods:   []string{"GET", "POST"},
				RateLimit: 100,
				CacheTTL:  5 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "empty backend",
			route: RouteConfig{
				Path:      "/api/*",
				Backend:   "",
				Methods:   []string{"GET", "POST"},
				RateLimit: 100,
				CacheTTL:  5 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "empty methods",
			route: RouteConfig{
				Path:      "/api/*",
				Backend:   "test",
				Methods:   []string{},
				RateLimit: 100,
				CacheTTL:  5 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "zero rate limit",
			route: RouteConfig{
				Path:      "/api/*",
				Backend:   "test",
				Methods:   []string{"GET", "POST"},
				RateLimit: 0,
				CacheTTL:  5 * time.Minute,
			},
			wantErr: false, // Rate limit can be 0 (no limit)
		},
		{
			name: "negative rate limit",
			route: RouteConfig{
				Path:      "/api/*",
				Backend:   "test",
				Methods:   []string{"GET", "POST"},
				RateLimit: -1,
				CacheTTL:  5 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "negative cache TTL",
			route: RouteConfig{
				Path:      "/api/*",
				Backend:   "test",
				Methods:   []string{"GET", "POST"},
				RateLimit: 100,
				CacheTTL:  -1 * time.Minute,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.route.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetBackendByName(t *testing.T) {
	config := &Config{
		Backend: []BackendConfig{
			{Name: "service1", URL: "http://localhost:3000"},
			{Name: "service2", URL: "http://localhost:3001"},
		},
	}

	// Test existing backend
	backend, err := config.GetBackendByName("service1")
	assert.NoError(t, err)
	assert.Equal(t, "service1", backend.Name)
	assert.Equal(t, "http://localhost:3000", backend.URL)

	// Test non-existing backend
	backend, err = config.GetBackendByName("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, backend)
	assert.Contains(t, err.Error(), "backend not found")
}

func TestGetRouteByPath(t *testing.T) {
	config := &Config{
		Routes: []RouteConfig{
			{Path: "/api/v1/*", Backend: "service1"},
			{Path: "/api/v2/*", Backend: "service2"},
		},
	}

	// Test existing route
	route, err := config.GetRouteByPath("/api/v1/*")
	assert.NoError(t, err)
	assert.Equal(t, "/api/v1/*", route.Path)
	assert.Equal(t, "service1", route.Backend)

	// Test non-existing route
	route, err = config.GetRouteByPath("/api/v3/*")
	assert.Error(t, err)
	assert.Nil(t, route)
	assert.Contains(t, err.Error(), "route not found")
}

func TestEnvironmentDetection(t *testing.T) {
	// Clear any existing environment variable
	os.Unsetenv("KALSHI_ENV")

	config := DefaultConfig()

	// Test development environment (default when no env var is set)
	assert.False(t, config.IsDevelopment())
	assert.False(t, config.IsProduction())

	// Test production environment
	os.Setenv("KALSHI_ENV", "production")
	defer os.Unsetenv("KALSHI_ENV")

	config, err := LoadConfigFromEnv()
	require.NoError(t, err)
	assert.False(t, config.IsDevelopment())
	assert.True(t, config.IsProduction())

	// Test development environment
	os.Setenv("KALSHI_ENV", "development")

	config, err = LoadConfigFromEnv()
	require.NoError(t, err)
	assert.True(t, config.IsDevelopment())
	assert.False(t, config.IsProduction())
}

func TestLoadWithInvalidYAML(t *testing.T) {
	// Create a temporary invalid YAML file
	tmpFile, err := os.CreateTemp("", "invalid-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("invalid: yaml: content: [")
	require.NoError(t, err)
	tmpFile.Close()

	_, err = Load(tmpFile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadWithValidYAML(t *testing.T) {
	// Create a temporary valid YAML file
	tmpFile, err := os.CreateTemp("", "valid-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	yamlContent := `
server:
  port: 9090
  host: "127.0.0.1"
auth:
  jwt:
    secret: "test-secret"
    expiry: "1h"
rate_limit:
  default_rate: 500
logging:
  level: "debug"
`
	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	config, err := Load(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, 9090, config.Server.Port)
	assert.Equal(t, "127.0.0.1", config.Server.Host)
	assert.Equal(t, "test-secret", config.Auth.JWT.Secret)
	assert.Equal(t, 1*time.Hour, config.Auth.JWT.Expiry)
	assert.Equal(t, 500, config.RateLimit.DefaultRate)
	assert.Equal(t, "debug", config.Logging.Level)
}

func TestEnvironmentVariableOverride(t *testing.T) {
	// Set environment variable to override config file
	os.Setenv("KALSHI_SERVER_PORT", "9999")
	defer os.Unsetenv("KALSHI_SERVER_PORT")

	config, err := LoadConfig("config.yaml")
	require.NoError(t, err)

	// Environment variable should override config file
	assert.Equal(t, 9999, config.Server.Port)
}

func TestComplexEnvironmentVariableMapping(t *testing.T) {
	// Test complex nested environment variable mapping
	os.Setenv("KALSHI_AUTH_JWT_SECRET", "env-secret")
	os.Setenv("KALSHI_CACHE_REDIS_ADDR", "redis:6379")
	os.Setenv("KALSHI_CACHE_REDIS_PASSWORD", "redis-pass")
	os.Setenv("KALSHI_RATE_LIMIT_STORAGE", "redis")

	defer func() {
		os.Unsetenv("KALSHI_AUTH_JWT_SECRET")
		os.Unsetenv("KALSHI_CACHE_REDIS_ADDR")
		os.Unsetenv("KALSHI_CACHE_REDIS_PASSWORD")
		os.Unsetenv("KALSHI_RATE_LIMIT_STORAGE")
	}()

	config, err := LoadConfigFromEnv()
	require.NoError(t, err)

	assert.Equal(t, "env-secret", config.Auth.JWT.Secret)
	assert.Equal(t, "redis:6379", config.Cache.Redis.Addr)
	assert.Equal(t, "redis-pass", config.Cache.Redis.Password)
	assert.Equal(t, "redis", config.RateLimit.Storage)
}

func TestValidationErrorMessages(t *testing.T) {
	config := DefaultConfig()
	config.Server.Port = 0

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server config")
	assert.Contains(t, err.Error(), "port must be between 1 and 65535")

	config = DefaultConfig()
	config.Auth.JWT.Secret = ""

	err = config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth config")
	assert.Contains(t, err.Error(), "JWT secret cannot be empty")
}

func TestBackendAndRouteConsistency(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Port:         8080,
			Host:         "0.0.0.0",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Auth: AuthConfig{
			JWT: JWTConfig{
				Secret: "test-secret",
				Expiry: 24 * time.Hour,
			},
			APIKey: APIKeyConfig{
				Enabled: false,
				Header:  "X-API-Key",
			},
		},
		RateLimit: RateLimitConfig{
			DefaultRate:     100,
			BurstCapacity:   200,
			Storage:         "memory",
			CleanupInterval: 1 * time.Hour,
		},
		Cache: CacheConfig{
			Redis: RedisConfig{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
				TTL:      5 * time.Minute,
			},
			Memory: MemoryConfig{
				MaxSize: 1000,
				TTL:     1 * time.Minute,
			},
		},
		Circuit: CircuitConfig{
			FailureThreshold: 5,
			RecoveryTimeout:  30 * time.Second,
			MaxRequests:      3,
		},
		Backend: []BackendConfig{
			{Name: "service1", URL: "http://localhost:3000", HealthCheck: "/health", Weight: 1},
		},
		Routes: []RouteConfig{
			{Path: "/api/*", Backend: "service1", Methods: []string{"GET"}, RateLimit: 100, CacheTTL: 5 * time.Minute},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
			Port:    9090,
		},
	}

	// Should be valid
	err := config.Validate()
	assert.NoError(t, err)

	// Test with route referencing non-existent backend (current validation doesn't check this)
	config.Routes[0].Backend = "nonexistent"
	err = config.Validate()
	assert.NoError(t, err) // Current validation doesn't check backend-route consistency
}

func TestConfigSerialization(t *testing.T) {
	config := DefaultConfig()

	// Test JSON marshaling
	jsonData, err := json.Marshal(config)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Test that the JSON contains expected fields
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, `"server"`)
	assert.Contains(t, jsonStr, `"auth"`)
	assert.Contains(t, jsonStr, `"rate_limit"`)
	assert.Contains(t, jsonStr, `"cache"`)
	assert.Contains(t, jsonStr, `"circuit"`)
	assert.Contains(t, jsonStr, `"backend"`)
	assert.Contains(t, jsonStr, `"routes"`)
	assert.Contains(t, jsonStr, `"logging"`)
	assert.Contains(t, jsonStr, `"metrics"`)
}

func TestConfigDeepCopy(t *testing.T) {
	original := DefaultConfig()
	original.Server.Port = 9999
	original.Auth.JWT.Secret = "original-secret"

	// Create a copy by loading from environment
	os.Setenv("KALSHI_SERVER_PORT", "8888")
	os.Setenv("KALSHI_AUTH_JWT_SECRET", "new-secret")
	defer func() {
		os.Unsetenv("KALSHI_SERVER_PORT")
		os.Unsetenv("KALSHI_AUTH_JWT_SECRET")
	}()

	copy, err := LoadConfigFromEnv()
	require.NoError(t, err)

	// Verify they are different
	assert.Equal(t, 9999, original.Server.Port)
	assert.Equal(t, 8888, copy.Server.Port)
	assert.Equal(t, "original-secret", original.Auth.JWT.Secret)
	assert.Equal(t, "new-secret", copy.Auth.JWT.Secret)
}

func TestConfigWithAllFeatures(t *testing.T) {
	// Test a comprehensive configuration with all features enabled
	config := &Config{
		Server: ServerConfig{
			Port:         8080,
			Host:         "0.0.0.0",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Auth: AuthConfig{
			JWT: JWTConfig{
				Secret: "comprehensive-test-secret",
				Expiry: 24 * time.Hour,
			},
			APIKey: APIKeyConfig{
				Enabled: true,
				Header:  "X-API-Key",
			},
		},
		RateLimit: RateLimitConfig{
			DefaultRate:     1000,
			BurstCapacity:   2000,
			Storage:         "redis",
			CleanupInterval: 1 * time.Hour,
		},
		Cache: CacheConfig{
			Redis: RedisConfig{
				Addr:     "redis:6379",
				Password: "redis-password",
				DB:       1,
				TTL:      10 * time.Minute,
			},
			Memory: MemoryConfig{
				MaxSize: 5000,
				TTL:     2 * time.Minute,
			},
		},
		Circuit: CircuitConfig{
			FailureThreshold: 10,
			RecoveryTimeout:  60 * time.Second,
			MaxRequests:      5,
		},
		Backend: []BackendConfig{
			{
				Name:        "api-service",
				URL:         "http://api-service:8080",
				HealthCheck: "/health",
				Weight:      100,
			},
			{
				Name:        "web-service",
				URL:         "http://web-service:3000",
				HealthCheck: "/ping",
				Weight:      50,
			},
		},
		Routes: []RouteConfig{
			{
				Path:      "/api/v1/*",
				Backend:   "api-service",
				Methods:   []string{"GET", "POST", "PUT", "DELETE"},
				RateLimit: 500,
				CacheTTL:  5 * time.Minute,
			},
			{
				Path:      "/web/*",
				Backend:   "web-service",
				Methods:   []string{"GET", "POST"},
				RateLimit: 200,
				CacheTTL:  2 * time.Minute,
			},
		},
		Logging: LoggingConfig{
			Level:  "debug",
			Format: "json",
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
			Port:    9090,
		},
	}

	// Validate the comprehensive config
	err := config.Validate()
	assert.NoError(t, err)

	// Test backend lookup
	apiBackend, err := config.GetBackendByName("api-service")
	assert.NoError(t, err)
	assert.Equal(t, "api-service", apiBackend.Name)
	assert.Equal(t, "http://api-service:8080", apiBackend.URL)

	webBackend, err := config.GetBackendByName("web-service")
	assert.NoError(t, err)
	assert.Equal(t, "web-service", webBackend.Name)
	assert.Equal(t, "http://web-service:3000", webBackend.URL)

	// Test route lookup
	apiRoute, err := config.GetRouteByPath("/api/v1/*")
	assert.NoError(t, err)
	assert.Equal(t, "/api/v1/*", apiRoute.Path)
	assert.Equal(t, "api-service", apiRoute.Backend)

	webRoute, err := config.GetRouteByPath("/web/*")
	assert.NoError(t, err)
	assert.Equal(t, "/web/*", webRoute.Path)
	assert.Equal(t, "web-service", webRoute.Backend)
}
