package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the main application configuration
// This struct defines all configuration options for the Kalshi API Gateway
// Configuration can be loaded from YAML files and environment variables
// Environment variables use the prefix KALSHI_ and replace dots with underscores
// Example: server.port becomes KALSHI_SERVER_PORT
type Config struct {
	Server    ServerConfig    `mapstructure:"server" json:"server"`         // HTTP server configuration
	Auth      AuthConfig      `mapstructure:"auth" json:"auth"`             // Authentication settings
	RateLimit RateLimitConfig `mapstructure:"rate_limit" json:"rate_limit"` // Rate limiting configuration
	Cache     CacheConfig     `mapstructure:"cache" json:"cache"`           // Caching configuration (Redis/Memory)
	Circuit   CircuitConfig   `mapstructure:"circuit" json:"circuit"`       // Circuit breaker configuration
	Backend   []BackendConfig `mapstructure:"backend" json:"backend"`       // Backend service definitions
	Routes    []RouteConfig   `mapstructure:"routes" json:"routes"`         // Route mapping configuration
	Logging   LoggingConfig   `mapstructure:"logging" json:"logging"`       // Logging configuration
	Metrics   MetricsConfig   `mapstructure:"metrics" json:"metrics"`       // Metrics/Prometheus configuration
}

// ServerConfig defines server-related configuration
type ServerConfig struct {
	Port         int           `mapstructure:"port" json:"port"`
	Host         string        `mapstructure:"host" json:"host"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" json:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" json:"idle_timeout"`
}

// AuthConfig defines authentication configuration
type AuthConfig struct {
	JWT    JWTConfig    `mapstructure:"jwt" json:"jwt"`
	APIKey APIKeyConfig `mapstructure:"api_key" json:"api_key"`
}

// JWTConfig defines JWT-specific configuration
type JWTConfig struct {
	Secret string        `mapstructure:"secret" json:"secret"`
	Expiry time.Duration `mapstructure:"expiry" json:"expiry"`
}

// APIKeyConfig defines API key authentication configuration
type APIKeyConfig struct {
	Enabled bool   `mapstructure:"enabled" json:"enabled"`
	Header  string `mapstructure:"header" json:"header"`
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	DefaultRate     int           `mapstructure:"default_rate" json:"default_rate"`
	BurstCapacity   int           `mapstructure:"burst_capacity" json:"burst_capacity"`
	Storage         string        `mapstructure:"storage" json:"storage"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval" json:"cleanup_interval"`
}

// CacheConfig defines caching configuration
type CacheConfig struct {
	Redis  RedisConfig  `mapstructure:"redis" json:"redis"`
	Memory MemoryConfig `mapstructure:"memory" json:"memory"`
}

// RedisConfig defines Redis-specific configuration
type RedisConfig struct {
	Addr     string        `mapstructure:"addr" json:"addr"`
	Password string        `mapstructure:"password" json:"password"`
	DB       int           `mapstructure:"db" json:"db"`
	TTL      time.Duration `mapstructure:"ttl" json:"ttl"`
}

// MemoryConfig defines in-memory cache configuration
type MemoryConfig struct {
	MaxSize int           `mapstructure:"max_size" json:"max_size"`
	TTL     time.Duration `mapstructure:"ttl" json:"ttl"`
}

// CircuitConfig defines circuit breaker configuration
type CircuitConfig struct {
	FailureThreshold int           `mapstructure:"failure_threshold" json:"failure_threshold"`
	RecoveryTimeout  time.Duration `mapstructure:"recovery_timeout" json:"recovery_timeout"`
	MaxRequests      int           `mapstructure:"max_requests" json:"max_requests"`
}

// BackendConfig defines backend service configuration
type BackendConfig struct {
	Name        string `mapstructure:"name" json:"name"`
	URL         string `mapstructure:"url" json:"url"`
	HealthCheck string `mapstructure:"health_check" json:"health_check"`
	Weight      int    `mapstructure:"weight" json:"weight"`
}

// RouteConfig defines route-specific configuration
type RouteConfig struct {
	Path      string        `mapstructure:"path" json:"path"`
	Backend   string        `mapstructure:"backend" json:"backend"`
	Methods   []string      `mapstructure:"methods" json:"methods"`
	RateLimit int           `mapstructure:"rate_limit" json:"rate_limit"`
	CacheTTL  time.Duration `mapstructure:"cache_ttl" json:"cache_ttl"`
}

// LoggingConfig defines logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level" json:"level"`
	Format string `mapstructure:"format" json:"format"`
}

// MetricsConfig defines metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled" json:"enabled"`
	Path    string `mapstructure:"path" json:"path"`
	Port    int    `mapstructure:"port" json:"port"`
}

// DefaultConfig returns a configuration with sensible defaults
// NOTE: For production deployment, override the JWT secret using environment variable KALSHI_AUTH_JWT_SECRET
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         8080,
			Host:         "0.0.0.0",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Auth: AuthConfig{
			JWT: JWTConfig{
				Secret: "your-secret-key-change-in-production",
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
			{
				Name:        "default",
				URL:         "http://localhost:3000",
				HealthCheck: "/health",
				Weight:      1,
			},
		},
		Routes: []RouteConfig{
			{
				Path:      "/api/*",
				Backend:   "default",
				Methods:   []string{"GET", "POST", "PUT", "DELETE"},
				RateLimit: 100,
				CacheTTL:  5 * time.Minute,
			},
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
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	// Use the Load function from loader.go
	return Load(configPath)
}

// LoadConfigFromEnv loads configuration from environment variables only
// Uses the same environment variable approach as Load() function
func LoadConfigFromEnv() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("KALSHI")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults first
	setDefaults()

	config := DefaultConfig()

	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config from environment variables: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server config: %w", err)
	}

	if err := c.Auth.Validate(); err != nil {
		return fmt.Errorf("auth config: %w", err)
	}

	if err := c.RateLimit.Validate(); err != nil {
		return fmt.Errorf("rate limit config: %w", err)
	}

	if err := c.Cache.Validate(); err != nil {
		return fmt.Errorf("cache config: %w", err)
	}

	if err := c.Circuit.Validate(); err != nil {
		return fmt.Errorf("circuit config: %w", err)
	}

	if err := c.Logging.Validate(); err != nil {
		return fmt.Errorf("logging config: %w", err)
	}

	if err := c.Metrics.Validate(); err != nil {
		return fmt.Errorf("metrics config: %w", err)
	}

	// Validate backends
	for i, backend := range c.Backend {
		if err := backend.Validate(); err != nil {
			return fmt.Errorf("backend[%d]: %w", i, err)
		}
	}

	// Validate routes
	for i, route := range c.Routes {
		if err := route.Validate(); err != nil {
			return fmt.Errorf("route[%d]: %w", i, err)
		}
	}

	return nil
}

// Validate validates server configuration
func (s *ServerConfig) Validate() error {
	if s.Port < 1 || s.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", s.Port)
	}

	if s.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if s.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}

	if s.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}

	if s.IdleTimeout <= 0 {
		return fmt.Errorf("idle timeout must be positive")
	}

	return nil
}

// Validate validates authentication configuration
func (a *AuthConfig) Validate() error {
	if err := a.JWT.Validate(); err != nil {
		return fmt.Errorf("jwt: %w", err)
	}

	if err := a.APIKey.Validate(); err != nil {
		return fmt.Errorf("api_key: %w", err)
	}

	return nil
}

// Validate validates JWT configuration
func (j *JWTConfig) Validate() error {
	if j.Secret == "" {
		return fmt.Errorf("JWT secret cannot be empty - set KALSHI_AUTH_JWT_SECRET environment variable")
	}

	if j.Expiry <= 0 {
		return fmt.Errorf("JWT expiry must be positive, got %v", j.Expiry)
	}

	return nil
}

// Validate validates API key configuration
func (ak *APIKeyConfig) Validate() error {
	if ak.Enabled && ak.Header == "" {
		return fmt.Errorf("header cannot be empty when API key auth is enabled")
	}

	return nil
}

// Validate validates rate limit configuration
func (r *RateLimitConfig) Validate() error {
	if r.DefaultRate <= 0 {
		return fmt.Errorf("rate limit default rate must be positive, got %d", r.DefaultRate)
	}

	if r.BurstCapacity <= 0 {
		return fmt.Errorf("rate limit burst capacity must be positive, got %d", r.BurstCapacity)
	}

	if r.Storage != "memory" && r.Storage != "redis" {
		return fmt.Errorf("rate limit storage must be 'memory' or 'redis', got %s", r.Storage)
	}

	if r.CleanupInterval <= 0 {
		return fmt.Errorf("rate limit cleanup interval must be positive, got %v", r.CleanupInterval)
	}

	return nil
}

// Validate validates cache configuration
func (c *CacheConfig) Validate() error {
	if err := c.Redis.Validate(); err != nil {
		return fmt.Errorf("redis: %w", err)
	}

	if err := c.Memory.Validate(); err != nil {
		return fmt.Errorf("memory: %w", err)
	}

	return nil
}

// Validate validates Redis configuration
func (r *RedisConfig) Validate() error {
	if r.Addr == "" {
		return fmt.Errorf("addr cannot be empty")
	}

	if r.DB < 0 || r.DB > 15 {
		return fmt.Errorf("db must be between 0 and 15, got %d", r.DB)
	}

	if r.TTL <= 0 {
		return fmt.Errorf("ttl must be positive")
	}

	return nil
}

// Validate validates memory cache configuration
func (m *MemoryConfig) Validate() error {
	if m.MaxSize <= 0 {
		return fmt.Errorf("max size must be positive")
	}

	if m.TTL <= 0 {
		return fmt.Errorf("ttl must be positive")
	}

	return nil
}

// Validate validates circuit breaker configuration
func (c *CircuitConfig) Validate() error {
	if c.FailureThreshold <= 0 {
		return fmt.Errorf("failure threshold must be positive")
	}

	if c.RecoveryTimeout <= 0 {
		return fmt.Errorf("recovery timeout must be positive")
	}

	if c.MaxRequests <= 0 {
		return fmt.Errorf("max requests must be positive")
	}

	return nil
}

// Validate validates backend configuration
func (b *BackendConfig) Validate() error {
	if b.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if b.URL == "" {
		return fmt.Errorf("url cannot be empty")
	}

	if b.Weight <= 0 {
		return fmt.Errorf("weight must be positive")
	}

	return nil
}

// Validate validates route configuration
func (r *RouteConfig) Validate() error {
	if r.Path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if r.Backend == "" {
		return fmt.Errorf("backend cannot be empty")
	}

	if len(r.Methods) == 0 {
		return fmt.Errorf("methods cannot be empty")
	}

	if r.RateLimit < 0 {
		return fmt.Errorf("rate limit cannot be negative")
	}

	if r.CacheTTL < 0 {
		return fmt.Errorf("cache ttl cannot be negative")
	}

	return nil
}

// Validate validates logging configuration
func (l *LoggingConfig) Validate() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}

	if !validLevels[l.Level] {
		return fmt.Errorf("invalid log level: %s", l.Level)
	}

	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}

	if !validFormats[l.Format] {
		return fmt.Errorf("invalid log format: %s", l.Format)
	}

	return nil
}

// Validate validates metrics configuration
func (m *MetricsConfig) Validate() error {
	if m.Enabled {
		if m.Port < 1 || m.Port > 65535 {
			return fmt.Errorf("metrics port must be between 1 and 65535, got %d", m.Port)
		}

		if m.Path == "" {
			return fmt.Errorf("metrics path cannot be empty")
		}
	}

	return nil
}

// GetBackendByName returns a backend configuration by name
func (c *Config) GetBackendByName(name string) (*BackendConfig, error) {
	for _, backend := range c.Backend {
		if backend.Name == name {
			return &backend, nil
		}
	}
	return nil, fmt.Errorf("backend not found: %s", name)
}

// GetRouteByPath returns a route configuration by path
func (c *Config) GetRouteByPath(path string) (*RouteConfig, error) {
	for _, route := range c.Routes {
		if route.Path == path {
			return &route, nil
		}
	}
	return nil, fmt.Errorf("route not found: %s", path)
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return os.Getenv("KALSHI_ENV") == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return os.Getenv("KALSHI_ENV") == "production"
}
