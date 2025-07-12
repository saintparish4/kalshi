package testing

import (
	"errors"
	"testing"
	"time"

	"kalshi/pkg/logger"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		format  string
		wantErr bool
	}{
		{"valid debug json", "debug", "json", false},
		{"valid info console", "info", "console", false},
		{"valid warn json", "warn", "json", false},
		{"valid error console", "error", "console", false},
		{"invalid level", "invalid", "json", true},
		{"invalid format", "info", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerInstance, err := logger.New(tt.level, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && loggerInstance == nil {
				t.Error("New() returned nil logger when no error expected")
			}
			if loggerInstance != nil {
				loggerInstance.Close()
			}
		})
	}
}

func TestLogger_WithFields(t *testing.T) {
	loggerInstance, err := logger.New("info", "console")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer loggerInstance.Close()

	fields := map[string]interface{}{
		"user_id":    123,
		"request_id": "req-123",
		"ip":         "192.168.1.1",
	}

	loggerWithFields := loggerInstance.WithFields(fields)
	if loggerWithFields == nil {
		t.Error("WithFields() returned nil logger")
	}

	// Test that the logger with fields works
	loggerWithFields.Info("Test message")
}

func TestLogger_WithError(t *testing.T) {
	loggerInstance, err := logger.New("error", "console")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer loggerInstance.Close()

	testErr := errors.New("test error")
	loggerWithError := loggerInstance.WithError(testErr)
	if loggerWithError == nil {
		t.Error("WithError() returned nil logger")
	}

	// Test that the logger with error works
	loggerWithError.Error("Test error message")
}

func TestLogger_Close(t *testing.T) {
	loggerInstance, err := logger.New("info", "console")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test that Close() doesn't panic
	err = loggerInstance.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Test that Close() on nil logger doesn't panic
	var nilLogger *logger.Logger
	err = nilLogger.Close()
	if err != nil {
		t.Errorf("Close() on nil logger returned error: %v", err)
	}
}

func TestInitGlobal(t *testing.T) {
	// Test valid initialization
	err := logger.InitGlobal("info", "console")
	if err != nil {
		t.Errorf("InitGlobal() failed: %v", err)
	}

	// Test that global functions work after initialization
	logger.Info("Test info message")
	logger.Debug("Test debug message")
	logger.Warn("Test warn message")
	logger.Error("Test error message")

	// Test invalid initialization
	err = logger.InitGlobal("invalid", "console")
	if err == nil {
		t.Error("InitGlobal() should fail with invalid level")
	}
}

func TestGlobalFunctions_WithoutInit(t *testing.T) {
	// These should not panic even without initialization
	logger.Info("Test without init")
	logger.Debug("Test without init")
	logger.Warn("Test without init")
	logger.Error("Test without init")
}

func TestConcurrentUsage(t *testing.T) {
	// Initialize global logger
	err := logger.InitGlobal("info", "console")
	if err != nil {
		t.Fatalf("Failed to initialize global logger: %v", err)
	}

	// Test concurrent usage
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Info("Concurrent message", "goroutine_id", id)
			logger.Error("Concurrent error", "goroutine_id", id)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestStructuredLogging(t *testing.T) {
	loggerInstance, err := logger.New("info", "json")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer loggerInstance.Close()

	// Test structured logging with various field types
	loggerInstance.Info("Structured log message",
		"string_field", "value",
		"int_field", 123,
		"float_field", 45.67,
		"bool_field", true,
		"time_field", time.Now(),
	)
}

func TestLogLevels(t *testing.T) {
	loggerInstance, err := logger.New("debug", "console")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer loggerInstance.Close()

	// Test all log levels
	loggerInstance.Debug("Debug message")
	loggerInstance.Info("Info message")
	loggerInstance.Warn("Warn message")
	loggerInstance.Error("Error message")
}

func BenchmarkLogger_Info(b *testing.B) {
	loggerInstance, err := logger.New("info", "json")
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer loggerInstance.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loggerInstance.Info("Benchmark message", "iteration", i)
	}
}

func BenchmarkGlobalLogger_Info(b *testing.B) {
	err := logger.InitGlobal("info", "json")
	if err != nil {
		b.Fatalf("Failed to initialize global logger: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message", "iteration", i)
	}
}
