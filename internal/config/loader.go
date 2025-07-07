package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Load loads configuration from file and environment variables
// Environment variables can override config file values using the format:
// KALSHI_SERVER_PORT, KALSHI_AUTH_JWT_SECRET, etc.
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	// Set environment variable prefix and key replacer for consistent naming
	viper.SetEnvPrefix("KALSHI")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set Default
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", configPath, err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config from file '%s': %w", configPath, err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
// NOTE: For production, change the JWT secret via environment variable KALSHI_AUTH_JWT_SECRET
func setDefaults() {
	// Server defaults - HTTP server configuration
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")

	// Auth Defaults - Authentication and authorization settings
	viper.SetDefault("auth.jwt.secret", "your-secret-key-change-this-in-production")
	viper.SetDefault("auth.jwt.expiry", "24h")
	viper.SetDefault("auth.api_key.enabled", true)
	viper.SetDefault("auth.api_key.header", "X-API-Key")

	// Rate Limit Defaults - Request rate limiting configuration
	viper.SetDefault("rate_limit.default_rate", 100)   // requests per second
	viper.SetDefault("rate_limit.burst_capacity", 200) // burst capacity
	viper.SetDefault("rate_limit.storage", "memory")   // memory or redis
	viper.SetDefault("rate_limit.cleanup_interval", "60s")

	// Cache Defaults - Caching configuration for Redis and memory
	viper.SetDefault("cache.redis.addr", "localhost:6379")
	viper.SetDefault("cache.redis.db", 0)
	viper.SetDefault("cache.redis.ttl", "300s")
	viper.SetDefault("cache.memory.max_size", 1000)
	viper.SetDefault("cache.memory.ttl", "60s")

	// Circuit Breaker Defaults - Circuit breaker pattern configuration
	viper.SetDefault("circuit.failure_threshold", 5)    // failures before opening circuit
	viper.SetDefault("circuit.recovery_timeout", "30s") // time to wait before attempting recovery
	viper.SetDefault("circuit.max_requests", 3)         // max requests when half-open

	// Logging Defaults - Application logging configuration
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

	// Metrics Defaults - Prometheus metrics configuration
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.path", "/metrics")
	viper.SetDefault("metrics.port", 9090)
}
