# Performance Test Tool with Authentication

This performance test tool supports testing the Kalshi API Gateway with various authentication methods.

## Features

- **Multiple Authentication Types**: JWT, API Key, or both
- **Configurable Test Parameters**: RPS, duration, concurrency, etc.
- **Environment Variable Support**: Override settings via environment variables
- **Detailed Results**: Latency statistics, error rates, memory usage
- **JSON Output**: Structured results for analysis

## Usage

### Basic Test (No Authentication)

```bash
./cmd/performance/performance_test \
    --base-url=http://localhost:8080 \
    --duration=1m \
    --initial-rps=50 \
    --max-rps=500
```

### JWT Authentication

```bash
./cmd/performance/performance_test \
    --auth=true \
    --auth-type=jwt \
    --jwt-secret="your-secret-key" \
    --username="test-user" \
    --role="user" \
    --base-url=http://localhost:8080 \
    --duration=1m
```

### API Key Authentication

```bash
./cmd/performance/performance_test \
    --auth=true \
    --auth-type=api-key \
    --api-key="your-api-key" \
    --api-key-header="X-API-Key" \
    --base-url=http://localhost:8080 \
    --duration=1m
```

### Both Authentication Methods

```bash
./cmd/performance/performance_test \
    --auth=true \
    --auth-type=both \
    --jwt-secret="your-secret-key" \
    --api-key="your-api-key" \
    --username="test-user" \
    --role="user" \
    --base-url=http://localhost:8080 \
    --duration=1m
```

## Command Line Options

### Authentication Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--auth` | bool | false | Enable authentication for tests |
| `--auth-type` | string | "jwt" | Authentication type: jwt, api-key, or both |
| `--jwt-secret` | string | "your-secret-key..." | JWT secret for token generation |
| `--api-key` | string | "" | API key for authentication |
| `--api-key-header` | string | "X-API-Key" | API key header name |
| `--username` | string | "test-user" | Username for JWT token generation |
| `--role` | string | "user" | User role for JWT token generation |

### Performance Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--base-url` | string | "http://localhost:8080" | Base URL for the API Gateway |
| `--duration` | duration | "1m" | Test duration |
| `--initial-rps` | int | 50 | Initial requests per second |
| `--max-rps` | int | 500 | Maximum requests per second |
| `--ramp-up` | duration | "30s" | Ramp-up duration |
| `--warmup` | duration | "10s" | Warmup duration |
| `--concurrency` | int | 100 | Maximum concurrent connections |
| `--workers` | int | 50 | Worker pool size |

### Output Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | "" | Output file for results (optional) |
| `--verbose` | bool | false | Enable verbose output |

## Environment Variables

You can override settings using environment variables:

| Variable | Description |
|----------|-------------|
| `TEST_BASE_URL` | Base URL for the API Gateway |
| `TEST_MAX_RPS` | Maximum requests per second |
| `TEST_WORKER_POOL` | Worker pool size |
| `TEST_API_KEY` | API key for authentication |
| `TEST_JWT_SECRET` | JWT secret for token generation |

## Example Script

Run the included script to test all authentication methods:

```bash
./scripts/run_auth_performance_tests.sh
```

This script will run tests with:
1. No authentication
2. JWT authentication
3. API Key authentication
4. Both authentication methods

## Output Format

The tool generates JSON results with the following structure:

```json
{
  "total_requests": 1000,
  "successful_requests": 950,
  "failed_requests": 50,
  "avg_latency": "150ms",
  "p95_latency": "250ms",
  "p99_latency": "350ms",
  "min_latency": "50ms",
  "max_latency": "500ms",
  "actual_rps": 95.5,
  "error_rate": 5.0,
  "test_duration": "1m",
  "memory_usage": {
    "heap_alloc_mb": 25,
    "heap_sys_mb": 50,
    "num_goroutines": 100
  },
  "concurrent_connections": 50,
  "response_codes": {
    "200": 900,
    "401": 50
  },
  "performance_tier": "Good"
}
```

## Authentication Details

### JWT Authentication

- Generates JWT tokens with configurable username and role
- Uses the same secret as the API Gateway
- Tokens are valid for 1 hour by default
- Sent in `Authorization: Bearer <token>` header

### API Key Authentication

- Uses provided API key directly
- Configurable header name (default: `X-API-Key`)
- Can be used with query parameter fallback
- Supports custom rate limits per key

### Both Authentication Methods

- Tries JWT first, then API key
- Useful for testing endpoints that support multiple auth methods
- Provides fallback authentication options

## Integration with API Gateway

The performance test tool is designed to work with the Kalshi API Gateway's authentication system:

- **JWT**: Compatible with the gateway's JWT middleware
- **API Key**: Works with the gateway's API key validation
- **Optional Auth**: Tests endpoints that support both authenticated and anonymous access
- **Required Auth**: Tests endpoints that require valid authentication

## Troubleshooting

### Common Issues

1. **401 Unauthorized**: Check that authentication credentials are correct
2. **Connection Refused**: Ensure the API Gateway is running on the specified URL
3. **High Error Rate**: May indicate authentication issues or server overload
4. **Low RPS**: Check network connectivity and server capacity

### Debug Mode

Enable verbose output to see detailed configuration and request information:

```bash
./cmd/performance/performance_test --verbose --auth=true --auth-type=jwt
``` 