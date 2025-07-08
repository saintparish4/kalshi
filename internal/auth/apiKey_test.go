package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPIKeyManager(t *testing.T) {
	mockStorage := &MockStorage{}
	manager := NewAPIKeyManager(mockStorage)

	assert.NotNil(t, manager)
	assert.Equal(t, mockStorage, manager.storage)
}

func TestGenerateSecureAPIKey(t *testing.T) {
	key1 := generateSecureAPIKey()
	key2 := generateSecureAPIKey()

	// Keys should be different
	assert.NotEqual(t, key1, key2)

	// Keys should have the correct format
	assert.Contains(t, key1, "kalshi_")
	assert.Contains(t, key2, "kalshi_")

	// Keys should be 71 characters long (32 bytes = 64 hex chars + prefix)
	assert.Len(t, key1, 71) // "kalshi_" + 64 hex chars
	assert.Len(t, key2, 71)
}

func TestAPIKeyManager_CreateAPIKey(t *testing.T) {
	mockStorage := NewMockStorage()
	manager := NewAPIKeyManager(mockStorage)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		apiKey, err := manager.CreateAPIKey(ctx, "user123", 1000, "Test key", 24*time.Hour)
		require.NoError(t, err)
		assert.NotEmpty(t, apiKey)
		assert.Contains(t, apiKey, "kalshi_")

		// Verify key was stored
		info, err := manager.ValidateAPIKey(ctx, apiKey)
		require.NoError(t, err)
		assert.Equal(t, "user123", info.UserID)
		assert.Equal(t, 1000, info.RateLimit)
		assert.True(t, info.Enabled)
		assert.Equal(t, "Test key", info.Description)
		assert.True(t, info.CreatedAt.After(time.Now().Add(-time.Minute)))
		assert.True(t, info.ExpiresAt.After(time.Now().Add(23*time.Hour)))
	})

	t.Run("empty user ID", func(t *testing.T) {
		_, err := manager.CreateAPIKey(ctx, "", 1000, "Test key", 24*time.Hour)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidUserID, err)
	})

	t.Run("negative rate limit", func(t *testing.T) {
		_, err := manager.CreateAPIKey(ctx, "user123", -1, "Test key", 24*time.Hour)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidRateLimit, err)
	})

	t.Run("no expiration", func(t *testing.T) {
		apiKey, err := manager.CreateAPIKey(ctx, "user456", 500, "No expiration", 0)
		require.NoError(t, err)

		info, err := manager.ValidateAPIKey(ctx, apiKey)
		require.NoError(t, err)
		assert.True(t, info.ExpiresAt.IsZero())
	})
}

func TestAPIKeyManager_ValidateAPIKey(t *testing.T) {
	mockStorage := NewMockStorage()
	manager := NewAPIKeyManager(mockStorage)
	ctx := context.Background()

	// Create a test key
	apiKey, err := manager.CreateAPIKey(ctx, "user123", 1000, "Test key", 24*time.Hour)
	require.NoError(t, err)

	t.Run("valid key", func(t *testing.T) {
		info, err := manager.ValidateAPIKey(ctx, apiKey)
		require.NoError(t, err)
		assert.Equal(t, "user123", info.UserID)
		assert.Equal(t, 1000, info.RateLimit)
		assert.True(t, info.Enabled)
	})

	t.Run("empty key", func(t *testing.T) {
		_, err := manager.ValidateAPIKey(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAPIKey, err)
	})

	t.Run("non-existent key", func(t *testing.T) {
		_, err := manager.ValidateAPIKey(ctx, "kalshi_nonexistentkey")
		assert.Error(t, err)
		assert.Equal(t, ErrAPIKeyNotFound, err)
	})

	t.Run("disabled key", func(t *testing.T) {
		// Revoke the key
		err := manager.RevokeAPIKey(ctx, apiKey)
		require.NoError(t, err)

		// Try to validate it
		_, err = manager.ValidateAPIKey(ctx, apiKey)
		assert.Error(t, err)
		assert.Equal(t, ErrAPIKeyDisabled, err)
	})

	t.Run("expired key", func(t *testing.T) {
		// Create a key that expires in 1 millisecond
		expiredKey, err := manager.CreateAPIKey(ctx, "user456", 500, "Expired key", time.Millisecond)
		require.NoError(t, err)

		// Wait for it to expire
		time.Sleep(10 * time.Millisecond)

		_, err = manager.ValidateAPIKey(ctx, expiredKey)
		assert.Error(t, err)
		assert.Equal(t, ErrAPIKeyExpired, err)
	})
}

func TestAPIKeyManager_RevokeAPIKey(t *testing.T) {
	mockStorage := NewMockStorage()
	manager := NewAPIKeyManager(mockStorage)
	ctx := context.Background()

	// Create a test key
	apiKey, err := manager.CreateAPIKey(ctx, "user123", 1000, "Test key", 24*time.Hour)
	require.NoError(t, err)

	t.Run("successful revocation", func(t *testing.T) {
		err := manager.RevokeAPIKey(ctx, apiKey)
		assert.NoError(t, err)

		// Verify key is disabled
		_, err = manager.ValidateAPIKey(ctx, apiKey)
		assert.Error(t, err)
		assert.Equal(t, ErrAPIKeyDisabled, err)
	})

	t.Run("empty key", func(t *testing.T) {
		err := manager.RevokeAPIKey(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAPIKey, err)
	})

	t.Run("non-existent key", func(t *testing.T) {
		err := manager.RevokeAPIKey(ctx, "kalshi_nonexistentkey")
		assert.Error(t, err)
		assert.Equal(t, ErrAPIKeyNotFound, err)
	})
}

func TestAPIKeyManager_ListAPIKeys(t *testing.T) {
	mockStorage := NewMockStorage()
	manager := NewAPIKeyManager(mockStorage)
	ctx := context.Background()

	t.Run("empty list", func(t *testing.T) {
		list, err := manager.ListAPIKeys(ctx, "user123")
		require.NoError(t, err)
		assert.Equal(t, "user123", list.UserID)
		assert.Empty(t, list.Keys)
	})

	t.Run("with keys", func(t *testing.T) {
		// Create multiple keys for the same user
		_, err := manager.CreateAPIKey(ctx, "user456", 1000, "Key 1", 24*time.Hour)
		require.NoError(t, err)
		_, err = manager.CreateAPIKey(ctx, "user456", 500, "Key 2", 12*time.Hour)
		require.NoError(t, err)

		list, err := manager.ListAPIKeys(ctx, "user456")
		require.NoError(t, err)
		assert.Equal(t, "user456", list.UserID)
		assert.Len(t, list.Keys, 2)

		// Verify both keys are in the list
		keyIDs := make(map[string]bool)
		for _, key := range list.Keys {
			keyIDs[key.Description] = true
		}
		assert.True(t, keyIDs["Key 1"])
		assert.True(t, keyIDs["Key 2"])
	})

	t.Run("empty user ID", func(t *testing.T) {
		_, err := manager.ListAPIKeys(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidUserID, err)
	})
}

func TestAPIKeyManager_UpdateAPIKey(t *testing.T) {
	mockStorage := NewMockStorage()
	manager := NewAPIKeyManager(mockStorage)
	ctx := context.Background()

	// Create a test key
	apiKey, err := manager.CreateAPIKey(ctx, "user123", 1000, "Original description", 24*time.Hour)
	require.NoError(t, err)

	t.Run("update rate limit", func(t *testing.T) {
		updates := map[string]interface{}{
			"rate_limit": 2000,
		}
		err := manager.UpdateAPIKey(ctx, apiKey, updates)
		assert.NoError(t, err)

		// Verify update
		info, err := manager.ValidateAPIKey(ctx, apiKey)
		require.NoError(t, err)
		assert.Equal(t, 2000, info.RateLimit)
	})

	t.Run("update description", func(t *testing.T) {
		updates := map[string]interface{}{
			"description": "Updated description",
		}
		err := manager.UpdateAPIKey(ctx, apiKey, updates)
		assert.NoError(t, err)

		// Verify update
		info, err := manager.ValidateAPIKey(ctx, apiKey)
		require.NoError(t, err)
		assert.Equal(t, "Updated description", info.Description)
	})

	t.Run("disable key", func(t *testing.T) {
		updates := map[string]interface{}{
			"enabled": false,
		}
		err := manager.UpdateAPIKey(ctx, apiKey, updates)
		assert.NoError(t, err)

		// Verify key is disabled
		_, err = manager.ValidateAPIKey(ctx, apiKey)
		assert.Error(t, err)
		assert.Equal(t, ErrAPIKeyDisabled, err)
	})

	t.Run("empty key", func(t *testing.T) {
		updates := map[string]interface{}{
			"rate_limit": 3000,
		}
		err := manager.UpdateAPIKey(ctx, "", updates)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAPIKey, err)
	})

	t.Run("non-existent key", func(t *testing.T) {
		updates := map[string]interface{}{
			"rate_limit": 3000,
		}
		err := manager.UpdateAPIKey(ctx, "kalshi_nonexistentkey", updates)
		assert.Error(t, err)
		assert.Equal(t, ErrAPIKeyNotFound, err)
	})

	t.Run("invalid rate limit", func(t *testing.T) {
		updates := map[string]interface{}{
			"rate_limit": -1,
		}
		err := manager.UpdateAPIKey(ctx, apiKey, updates)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidRateLimit, err)
	})
}

func TestAPIKeyManager_DeleteAPIKey(t *testing.T) {
	mockStorage := NewMockStorage()
	manager := NewAPIKeyManager(mockStorage)
	ctx := context.Background()

	// Create a test key
	apiKey, err := manager.CreateAPIKey(ctx, "user123", 1000, "Test key", 24*time.Hour)
	require.NoError(t, err)

	t.Run("successful deletion", func(t *testing.T) {
		err := manager.DeleteAPIKey(ctx, apiKey)
		assert.NoError(t, err)

		// Verify key is deleted
		_, err = manager.ValidateAPIKey(ctx, apiKey)
		assert.Error(t, err)
		assert.Equal(t, ErrAPIKeyNotFound, err)
	})

	t.Run("empty key", func(t *testing.T) {
		err := manager.DeleteAPIKey(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAPIKey, err)
	})

	t.Run("non-existent key", func(t *testing.T) {
		err := manager.DeleteAPIKey(ctx, "kalshi_nonexistentkey")
		assert.Error(t, err)
		assert.Equal(t, ErrAPIKeyNotFound, err)
	})
}

func TestAPIKeyError(t *testing.T) {
	t.Run("error without cause", func(t *testing.T) {
		err := &APIKeyError{
			Type:    "test_error",
			Message: "test message",
		}
		errorStr := err.Error()
		assert.Contains(t, errorStr, "API Key test_error")
		assert.Contains(t, errorStr, "test message")
		assert.NotContains(t, errorStr, "cause:")
	})

	t.Run("error with cause", func(t *testing.T) {
		cause := &APIKeyError{Type: "cause", Message: "cause message"}
		err := &APIKeyError{
			Type:    "test_error",
			Message: "test message",
			Cause:   cause,
		}
		errorStr := err.Error()
		assert.Contains(t, errorStr, "API Key test_error")
		assert.Contains(t, errorStr, "test message")
		assert.Contains(t, errorStr, "cause:")
	})

	t.Run("unwrap", func(t *testing.T) {
		cause := &APIKeyError{Type: "cause", Message: "cause message"}
		err := &APIKeyError{
			Type:    "test_error",
			Message: "test message",
			Cause:   cause,
		}
		unwrapped := err.Unwrap()
		assert.Equal(t, cause, unwrapped)
	})
}

// MockStorage implements the storage.Storage interface for testing
type MockStorage struct {
	data map[string]string
	ttl  map[string]time.Time
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string]string),
		ttl:  make(map[string]time.Time),
	}
}

func (m *MockStorage) Get(ctx context.Context, key string) (string, error) {
	value, exists := m.data[key]
	if !exists {
		return "", &APIKeyError{Type: "not_found", Message: "key not found"}
	}

	// Return the value even if expired - let the API key validation handle expiration
	return value, nil
}

func (m *MockStorage) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	m.data[key] = value
	if ttl > 0 {
		m.ttl[key] = time.Now().Add(ttl)
	}
	return nil
}

func (m *MockStorage) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	delete(m.ttl, key)
	return nil
}

func (m *MockStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.data[key]
	return exists, nil
}

func (m *MockStorage) Increment(ctx context.Context, key string, by int64) (int64, error) {
	current, err := m.Get(ctx, key)
	if err != nil {
		// Key doesn't exist, create it
		newValue := by
		m.data[key] = string(rune(newValue))
		return newValue, nil
	}

	// Parse current value and increment
	var currentVal int64
	if _, err := fmt.Sscanf(current, "%d", &currentVal); err != nil {
		currentVal = 0
	}

	currentVal += by
	m.data[key] = string(rune(currentVal))
	return currentVal, nil
}

func (m *MockStorage) Close() error {
	return nil
}
