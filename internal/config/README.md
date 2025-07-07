# Configuration System

This package provides configuration management for the Kalshi API Gateway.

## Overview

The configuration system supports loading settings from:
- YAML configuration files
- Environment variables
- Default values

## Configuration Structure

### Server Configuration
```yaml
server:
  port: 8080                    # HTTP server port
  host: "0.0.0.0"              # Server host address
  read_timeout: "30s"           # Request read timeout
  write_timeout: "30s"          # Response write timeout
  idle_timeout: "60s"           # Connection idle timeout
```

### Authentication Configuration
```yaml
auth:
  jwt:
    secret: "your-secret-key"   # JWT signing secret (CHANGE IN PRODUCTION)
    expiry: "24h"               # JWT token expiry time
  api_key:
    enabled: true               # Enable API key authentication
    header: "X-API-Key"         # Header name for API key
```

### Rate Limiting Configuration
```yaml
rate_limit:
  default_rate: 100             # Requests per second
  burst_capacity: 200           # Burst capacity
  storage: "memory"             # Storage backend (memory/redis)
  cleanup_interval: "60s"       # Cleanup interval for expired entries
```

### Cache Configuration
```yaml
cache:
  redis:
    addr: "localhost:6379"      # Redis server address
    password: ""                # Redis password
    db: 0                       # Redis database number
    ttl: "300s"                 # Default TTL for cached items
  memory:
    max_size: 1000              # Maximum number of cached items
    ttl: "60s"                  # Default TTL for memory cache
```

### Circuit Breaker Configuration
```yaml
circuit:
  failure_threshold: 5          # Failures before opening circuit
  recovery_timeout: "30s"       # Time to wait before attempting recovery
  max_requests: 3               # Max requests when half-open
```

### Backend Services
```yaml
backend:
  - name: "service1"            # Backend service name
    url: "http://localhost:3000" # Backend service URL
    health_check: "/health"     # Health check endpoint
    weight: 100                 # Load balancing weight
```

### Route Configuration
```yaml
routes:
  - path: "/api/v1/*"           # Route path pattern
    backend: "service1"         # Backend service name
    methods: ["GET", "POST"]    # Allowed HTTP methods
    rate_limit: 500             # Route-specific rate limit
    cache_ttl: "120s"           # Route-specific cache TTL
```

### Logging Configuration
```yaml
logging:
  level: "info"                 # Log level (debug, info, warn, error)
  format: "json"                # Log format (json, text)
```

### Metrics Configuration
```yaml
metrics:
  enabled: true                 # Enable Prometheus metrics
  path: "/metrics"              # Metrics endpoint path
  port: 9090                    # Metrics server port
```

## Environment Variables

All configuration values can be overridden using environment variables with the `KALSHI_` prefix:

```bash
# Server configuration
export KALSHI_SERVER_PORT=8080
export KALSHI_SERVER_HOST=0.0.0.0

# Authentication
export KALSHI_AUTH_JWT_SECRET=your-production-secret
export KALSHI_AUTH_JWT_EXPIRY=24h

# Rate limiting
export KALSHI_RATE_LIMIT_DEFAULT_RATE=1000
export KALSHI_RATE_LIMIT_BURST_CAPACITY=200

# Cache
export KALSHI_CACHE_REDIS_ADDR=redis:6379
export KALSHI_CACHE_REDIS_PASSWORD=secret

# Logging
export KALSHI_LOGGING_LEVEL=debug
```

## Usage

### Loading from File
```go
config, err := config.Load("config.yaml")
if err != nil {
    log.Fatal(err)
}
```

### Loading from Environment Only
```go
config, err := config.LoadConfigFromEnv()
if err != nil {
    log.Fatal(err)
}
```

### Using Default Configuration
```go
config := config.DefaultConfig()
```

## Security Notes

⚠️ **IMPORTANT**: For production deployments:
1. Change the JWT secret using `KALSHI_AUTH_JWT_SECRET` environment variable
2. Use strong, randomly generated secrets
3. Never commit production secrets to version control
4. Consider using a secrets management service

## Validation

The configuration system includes comprehensive validation:
- Port numbers must be between 1-65535
- Timeouts must be positive values
- Rate limits must be positive
- Storage backends must be valid options
- Required fields cannot be empty

Validation errors provide clear, actionable messages to help resolve configuration issues. 