package testing

import (
	"context"
	"testing"
	"time"

	"kalshi/internal/storage"
)

func TestMemoryStorage_BasicOperations(t *testing.T) {
	storage := storage.NewMemoryStorage()
	defer storage.Close()
	ctx := context.Background()

	// Test Set and Get
	err := storage.Set(ctx, "test-key", "test-value", time.Hour)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	value, err := storage.Get(ctx, "test-key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if value != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", value)
	}

	// Test Exists
	exists, err := storage.Exists(ctx, "test-key")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Expected key to exist")
	}

	// Test Delete
	err = storage.Delete(ctx, "test-key")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	exists, err = storage.Exists(ctx, "test-key")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Expected key to not exist after deletion")
	}
}

func TestMemoryStorage_Increment(t *testing.T) {
	storage := storage.NewMemoryStorage()
	defer storage.Close()
	ctx := context.Background()

	// Test increment on new key
	value, err := storage.Increment(ctx, "counter", 5)
	if err != nil {
		t.Fatalf("Increment failed: %v", err)
	}
	if value != 5 {
		t.Errorf("Expected 5, got %d", value)
	}

	// Test increment on existing key
	value, err = storage.Increment(ctx, "counter", 3)
	if err != nil {
		t.Fatalf("Increment failed: %v", err)
	}
	if value != 8 {
		t.Errorf("Expected 8, got %d", value)
	}
}

func TestMemoryStorage_RateLimitOperations(t *testing.T) {
	storage := storage.NewMemoryStorage()
	defer storage.Close()
	ctx := context.Background()

	// Test SetTokens and GetTokens
	err := storage.SetTokens(ctx, "rate-limit", 10, time.Hour)
	if err != nil {
		t.Fatalf("SetTokens failed: %v", err)
	}

	tokens, err := storage.GetTokens(ctx, "rate-limit")
	if err != nil {
		t.Fatalf("GetTokens failed: %v", err)
	}
	if tokens != 10 {
		t.Errorf("Expected 10 tokens, got %d", tokens)
	}

	// Test DecrementTokens
	remaining, err := storage.DecrementTokens(ctx, "rate-limit")
	if err != nil {
		t.Fatalf("DecrementTokens failed: %v", err)
	}
	if remaining != 9 {
		t.Errorf("Expected 9 tokens remaining, got %d", remaining)
	}

	// Test IncrementTokens
	newTotal, err := storage.IncrementTokens(ctx, "rate-limit", 5)
	if err != nil {
		t.Fatalf("IncrementTokens failed: %v", err)
	}
	if newTotal != 14 {
		t.Errorf("Expected 14 tokens, got %d", newTotal)
	}
}

func TestMemoryStorage_Expiration(t *testing.T) {
	storage := storage.NewMemoryStorage()
	defer storage.Close()
	ctx := context.Background()

	// Set a key with very short TTL
	err := storage.Set(ctx, "expire-test", "value", 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Should exist immediately
	exists, err := storage.Exists(ctx, "expire-test")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Expected key to exist immediately")
	}

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Should not exist after expiration
	exists, err = storage.Exists(ctx, "expire-test")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Expected key to not exist after expiration")
	}
}

func TestMemoryStorage_InterfaceCompliance(t *testing.T) {
	// This test ensures the MemoryStorage implements both interfaces
	var _ storage.Storage = (*storage.MemoryStorage)(nil)
	var _ storage.RateLimitStorage = (*storage.MemoryStorage)(nil)
}
