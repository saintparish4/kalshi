# Logger Package

A structured logging package built on top of [zap](https://github.com/uber-go/zap) that provides both instance-based and global logging capabilities with support for different log levels and output formats.

## Features

- **Structured Logging**: Support for JSON and console output formats
- **Multiple Log Levels**: Debug, Info, Warn, Error, and Fatal levels
- **Thread-Safe**: Global logger functions are safe for concurrent use
- **Contextual Logging**: Add structured fields and error context to log messages
- **Configuration Validation**: Input validation with clear error messages
- **Performance**: Built on zap for high-performance logging

## Installation

```bash
go get kalshi/pkg/logger
```

## Quick Start

### Global Logger (Recommended for most applications)

```go
package main

import (
    "log"
    "kalshi/pkg/logger"
)

func main() {
    // Initialize global logger
    err := logger.InitGlobal("info", "json")
    if err != nil {
        log.Fatal("Failed to initialize logger:", err)
    }

    // Use global logging functions
    logger.Info("Application started")
    logger.Error("Something went wrong", "error", err)
}
```

### Instance-Based Logger

```go
package main

import (
    "log"
    "kalshi/pkg/logger"
)

func main() {
    // Create a new logger instance
    log, err := logger.New("debug", "console")
    if err != nil {
        log.Fatal("Failed to create logger:", err)
    }
    defer log.Close()

    // Use the logger instance
    log.Info("Processing request", "user_id", 123, "action", "login")
}
```

## Configuration

### Log Levels

- `debug`: Detailed debug information
- `info`: General information messages
- `warn`: Warning messages
- `error`: Error messages
- `fatal`: Fatal errors (exits the program)

### Log Formats

- `json`: Structured JSON output (recommended for production)
- `console`: Human-readable console output (recommended for development)

## Usage Examples

### Basic Logging

```go
logger.Info("User logged in", "user_id", 123, "ip", "192.168.1.1")
logger.Error("Database connection failed", "error", err)
logger.Debug("Processing request", "method", "POST", "path", "/api/users")
logger.Warn("High memory usage", "usage_percent", 85)
```

### Structured Logging with Fields

```go
// Create a logger with contextual fields
requestLogger := log.WithFields(map[string]interface{}{
    "request_id": "req-123",
    "user_id":    456,
    "ip":         "192.168.1.1",
})

// All subsequent log messages will include these fields
requestLogger.Info("Request started")
requestLogger.Info("Processing payment", "amount", 99.99, "currency", "USD")
requestLogger.Info("Request completed", "duration_ms", 150)
```

### Error Logging

```go
// Log with error context
errorLogger := log.WithError(err)
errorLogger.Error("Failed to process user request", "user_id", 789, "operation", "update")
```

### HTTP Middleware Example

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Log request start
        logger.Info("HTTP request started",
            "method", r.Method,
            "path", r.URL.Path,
            "user_id", getUserID(r),
            "ip", r.RemoteAddr,
        )

        // Process request
        next.ServeHTTP(w, r)

        // Log request completion
        duration := time.Since(start)
        logger.Info("HTTP request completed",
            "method", r.Method,
            "path", r.URL.Path,
            "user_id", getUserID(r),
            "ip", r.RemoteAddr,
            "duration_ms", duration.Milliseconds(),
            "status_code", getStatusCode(w),
        )
    })
}
```

## Best Practices

### 1. Initialize Logger Early

Initialize the global logger as early as possible in your application:

```go
func main() {
    // Initialize logger first
    if err := logger.InitGlobal("info", "json"); err != nil {
        log.Fatal("Failed to initialize logger:", err)
    }

    // Then initialize other components
    // ...
}
```

### 2. Use Structured Logging

Always include relevant context in your log messages:

```go
// Good
logger.Info("User authentication successful", "user_id", userID, "method", "oauth")

// Avoid
logger.Info("User logged in")
```

### 3. Handle Errors Properly

Use the `WithError` method for error logging:

```go
if err != nil {
    logger.Error("Failed to process request", "error", err, "user_id", userID)
    return err
}
```

### 4. Choose Appropriate Log Levels

- `debug`: Detailed information for debugging
- `info`: General application flow
- `warn`: Unexpected but handled situations
- `error`: Errors that don't stop the application
- `fatal`: Errors that require application shutdown

### 5. Use Contextual Loggers

Create contextual loggers for request-scoped logging:

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    requestLogger := log.WithFields(map[string]interface{}{
        "request_id": generateRequestID(),
        "user_id":    getUserID(r),
        "ip":         r.RemoteAddr,
    })

    requestLogger.Info("Request started")
    // ... process request
    requestLogger.Info("Request completed")
}
```

## Thread Safety

The global logger functions (`Info`, `Error`, `Debug`, `Warn`, `Fatal`) are thread-safe and can be called from multiple goroutines concurrently.

Instance-based loggers are also thread-safe as they wrap zap's thread-safe logger.

## Performance

This package is built on top of zap, which is one of the fastest structured logging libraries for Go. The wrapper adds minimal overhead while providing a simpler API.

## Error Handling

The package provides clear error messages for configuration issues:

```go
_, err := logger.New("invalid_level", "json")
// Error: invalid logger configuration: invalid log level: invalid_level (valid levels: debug, info, warn, error)

_, err = logger.New("info", "invalid_format")
// Error: invalid logger configuration: invalid log format: invalid_format (valid formats: json, console)
```

## Testing

Run the examples to see the logger in action:

```bash
go test -v ./pkg/logger -run Example
```

## Contributing

When contributing to this package:

1. Follow Go coding standards
2. Add tests for new features
3. Update documentation
4. Ensure thread safety for concurrent operations
5. Validate all inputs

## License

This package is part of the Kalshi project and follows the same license terms. 