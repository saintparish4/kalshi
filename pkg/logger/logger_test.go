package logger

import (
	"errors"
	"testing"
	"time"
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
			logger, err := New(tt.level, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && logger == nil {
				t.Error("New() returned nil logger when no error expected")
			}
			if logger != nil {
				logger.Close()
			}
		})
	}
}

func TestLogger_WithFields(t *testing.T) {
	logger, err := New("info", "console")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	fields := map[string]interface{}{
		"user_id":    123,
		"request_id": "req-123",
		"ip":         "192.168.1.1",
	}

	loggerWithFields := logger.WithFields(fields)
	if loggerWithFields == nil {
		t.Error("WithFields() returned nil logger")
	}

	// Test that the logger with fields works
	loggerWithFields.Info("Test message")
}

func TestLogger_WithError(t *testing.T) {
	logger, err := New("error", "console")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	testErr := errors.New("test error")
	loggerWithError := logger.WithError(testErr)
	if loggerWithError == nil {
		t.Error("WithError() returned nil logger")
	}

	// Test that the logger with error works
	loggerWithError.Error("Test error message")
}

func TestLogger_Close(t *testing.T) {
	logger, err := New("info", "console")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test that Close() doesn't panic
	err = logger.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Test that Close() on nil logger doesn't panic
	var nilLogger *Logger
	err = nilLogger.Close()
	if err != nil {
		t.Errorf("Close() on nil logger returned error: %v", err)
	}
}

func TestInitGlobal(t *testing.T) {
	// Test valid initialization
	err := InitGlobal("info", "console")
	if err != nil {
		t.Errorf("InitGlobal() failed: %v", err)
	}

	// Test that global functions work after initialization
	Info("Test info message")
	Debug("Test debug message")
	Warn("Test warn message")
	Error("Test error message")

	// Test invalid initialization
	err = InitGlobal("invalid", "console")
	if err == nil {
		t.Error("InitGlobal() should fail with invalid level")
	}
}

func TestGlobalFunctions_WithoutInit(t *testing.T) {
	// Reset global logger
	globalLogger = nil

	// These should not panic even without initialization
	Info("Test without init")
	Debug("Test without init")
	Warn("Test without init")
	Error("Test without init")
}

func TestConcurrentUsage(t *testing.T) {
	// Initialize global logger
	err := InitGlobal("info", "console")
	if err != nil {
		t.Fatalf("Failed to initialize global logger: %v", err)
	}

	// Test concurrent usage
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			Info("Concurrent message", "goroutine_id", id)
			Error("Concurrent error", "goroutine_id", id)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestStructuredLogging(t *testing.T) {
	logger, err := New("info", "json")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test structured logging with various field types
	logger.Info("Structured log message",
		"string_field", "value",
		"int_field", 123,
		"float_field", 45.67,
		"bool_field", true,
		"time_field", time.Now(),
	)
}

func TestLogLevels(t *testing.T) {
	logger, err := New("debug", "console")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test all log levels
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warn message")
	logger.Error("Error message")
}

func BenchmarkLogger_Info(b *testing.B) {
	logger, err := New("info", "json")
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message", "iteration", i)
	}
}

func BenchmarkGlobalLogger_Info(b *testing.B) {
	err := InitGlobal("info", "json")
	if err != nil {
		b.Fatalf("Failed to initialize global logger: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("Benchmark message", "iteration", i)
	}
}
