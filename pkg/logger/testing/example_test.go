package testing

import (
	"errors"
	"fmt"

	"kalshi/pkg/logger"
)

// Example demonstrates basic logger usage with global functions
func Example() {
	// Initialize the global logger
	err := logger.InitGlobal("info", "console")
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	// Basic logging
	logger.Info("Application started")
	logger.Debug("Debug information")
	logger.Warn("Warning message")
	logger.Error("Error occurred")
}

// ExampleNew demonstrates creating a new logger instance
func ExampleNew() {
	// Create a new logger instance
	log, err := logger.New("debug", "json")
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}
	defer log.Close()

	// Use the logger instance
	log.Info("Processing request", "user_id", 123, "action", "login")
	log.Debug("Request details", "method", "POST", "path", "/api/login")
}

// ExampleLogger_WithFields demonstrates structured logging with fields
func ExampleLogger_WithFields() {
	log, err := logger.New("info", "console")
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}
	defer log.Close()

	// Create a logger with structured fields
	requestLogger := log.WithFields(map[string]interface{}{
		"request_id": "req-123",
		"user_id":    456,
		"ip":         "192.168.1.1",
	})

	// Use the contextualized logger
	requestLogger.Info("Request started")
	requestLogger.Info("Processing payment", "amount", 99.99, "currency", "USD")
	requestLogger.Info("Request completed", "duration_ms", 150)
}

// ExampleLogger_WithError demonstrates error logging
func ExampleLogger_WithError() {
	log, err := logger.New("error", "console")
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}
	defer log.Close()

	// Simulate an error
	err = errors.New("database connection failed")

	// Log with error context
	errorLogger := log.WithError(err)
	errorLogger.Error("Failed to process user request", "user_id", 789, "operation", "update")
}
