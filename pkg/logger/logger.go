// Package logger provides a structured logging interface built on top of zap.
// It offers both instance-based and global logging capabilities with support
// for different log levels and output formats.
//
// Example usage:
//
//	// Initialize global logger
//	err := logger.InitGlobal("info", "json")
//	if err != nil {
//		log.Fatal("Failed to initialize logger:", err)
//	}
//
//	// Use global logging functions
//	logger.Info("Application started")
//	logger.Error("Something went wrong", "error", err)
//
//	// Or create instance-based logger
//	log, err := logger.New("debug", "console")
//	if err != nil {
//		log.Fatal("Failed to create logger:", err)
//	}
//	defer log.Close()
//
//	log.Info("Processing request", "user_id", 123, "action", "login")
package logger

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

// Logger wraps zap.SugaredLogger to provide a simplified interface
// with additional convenience methods for structured logging.
type Logger struct {
	*zap.SugaredLogger
}

// LogLevel represents the supported log levels
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
)

// LogFormat represents the supported log formats
type LogFormat string

const (
	JSONFormat    LogFormat = "json"
	ConsoleFormat LogFormat = "console"
)

// Global logger instance with thread safety
var (
	globalLogger *Logger
	loggerMutex  sync.RWMutex
)

// New creates a new Logger instance with the specified level and format.
// Returns an error if the configuration is invalid or logger creation fails.
func New(level string, format string) (*Logger, error) {
	if err := validateConfig(level, format); err != nil {
		return nil, fmt.Errorf("invalid logger configuration: %w", err)
	}

	var config zap.Config

	switch LogFormat(format) {
	case JSONFormat:
		config = zap.NewProductionConfig()
	default:
		config = zap.NewDevelopmentConfig()
	}

	// Set Log Level
	switch LogLevel(level) {
	case DebugLevel:
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case InfoLevel:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case WarnLevel:
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case ErrorLevel:
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	zapLogger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &Logger{zapLogger.Sugar()}, nil
}

// WithFields creates a new Logger instance with structured fields.
// This is useful for adding context to log messages.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	if l == nil {
		return nil
	}

	var args []interface{}
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{l.SugaredLogger.With(args...)}
}

// WithError creates a new Logger instance with an error field.
// This is a convenience method for logging errors with context.
func (l *Logger) WithError(err error) *Logger {
	if l == nil {
		return nil
	}
	return &Logger{l.SugaredLogger.With("error", err)}
}

// Close safely closes the logger and flushes any buffered log entries.
// Returns an error if flushing fails.
func (l *Logger) Close() error {
	if l == nil {
		return nil
	}
	return l.SugaredLogger.Sync()
}

// validateConfig validates the logger configuration parameters.
func validateConfig(level, format string) error {
	// Validate log level
	switch LogLevel(level) {
	case DebugLevel, InfoLevel, WarnLevel, ErrorLevel:
		// Valid level
	default:
		return fmt.Errorf("invalid log level: %s (valid levels: debug, info, warn, error)", level)
	}

	// Validate format
	switch LogFormat(format) {
	case JSONFormat, ConsoleFormat:
		// Valid format
	default:
		return fmt.Errorf("invalid log format: %s (valid formats: json, console)", format)
	}

	return nil
}

// InitGlobal initializes the global logger instance.
// This function is thread-safe and should be called once at application startup.
func InitGlobal(level string, format string) error {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	var err error
	globalLogger, err = New(level, format)
	if err != nil {
		return fmt.Errorf("failed to initialize global logger: %w", err)
	}
	return nil
}

// getGlobalLogger safely retrieves the global logger instance.
func getGlobalLogger() *Logger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	return globalLogger
}

// Info logs an info-level message using the global logger.
// This function is thread-safe.
func Info(args ...interface{}) {
	if logger := getGlobalLogger(); logger != nil {
		logger.Info(args...)
	}
}

// Error logs an error-level message using the global logger.
// This function is thread-safe.
func Error(args ...interface{}) {
	if logger := getGlobalLogger(); logger != nil {
		logger.Error(args...)
	}
}

// Debug logs a debug-level message using the global logger.
// This function is thread-safe.
func Debug(args ...interface{}) {
	if logger := getGlobalLogger(); logger != nil {
		logger.Debug(args...)
	}
}

// Warn logs a warning-level message using the global logger.
// This function is thread-safe.
func Warn(args ...interface{}) {
	if logger := getGlobalLogger(); logger != nil {
		logger.Warn(args...)
	}
}

// Fatal logs a fatal-level message using the global logger and exits the program.
// This function is thread-safe.
func Fatal(args ...interface{}) {
	if logger := getGlobalLogger(); logger != nil {
		logger.Fatal(args...)
	}
	os.Exit(1)
}
