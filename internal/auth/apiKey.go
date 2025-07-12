package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"kalshi/internal/storage"
	"time"
)

// APIKeyError represents API key-specific errors
type APIKeyError struct {
	Type    string
	Message string
	Cause   error
}

func (e *APIKeyError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("API Key %s: %s (cause: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("API Key %s: %s", e.Type, e.Message)
}

func (e *APIKeyError) Unwrap() error {
	return e.Cause
}

// Common API key errors
var (
	ErrInvalidAPIKey    = &APIKeyError{Type: "invalid_key", Message: "invalid API key"}
	ErrAPIKeyDisabled   = &APIKeyError{Type: "disabled", Message: "API key is disabled"}
	ErrAPIKeyExpired    = &APIKeyError{Type: "expired", Message: "API key has expired"}
	ErrAPIKeyNotFound   = &APIKeyError{Type: "not_found", Message: "API key not found"}
	ErrAPIKeyExists     = &APIKeyError{Type: "exists", Message: "API key already exists"}
	ErrInvalidUserID    = &APIKeyError{Type: "invalid_user", Message: "invalid user ID"}
	ErrInvalidRateLimit = &APIKeyError{Type: "invalid_rate_limit", Message: "invalid rate limit"}
)

type APIKeyManager struct {
	storage storage.Storage
}

type APIKeyInfo struct {
	UserID      string    `json:"user_id"`
	RateLimit   int       `json:"rate_limit"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsed    time.Time `json:"last_used,omitempty"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	Description string    `json:"description,omitempty"`
}

// APIKeyList represents a list of API keys for a user
type APIKeyList struct {
	UserID string       `json:"user_id"`
	Keys   []APIKeyInfo `json:"keys"`
}

func NewAPIKeyManager(storage storage.Storage) *APIKeyManager {
	return &APIKeyManager{
		storage: storage,
	}
}

// ValidateAPIKey validates an API key and returns its information
func (a *APIKeyManager) ValidateAPIKey(ctx context.Context, apiKey string) (*APIKeyInfo, error) {
	if apiKey == "" {
		return nil, ErrInvalidAPIKey
	}

	// Get API key info from storage
	key := "apikey:" + apiKey
	// Looking up API key in storage
	data, err := a.storage.Get(ctx, key)
	if err != nil {
		// Storage.Get failed for key
		return nil, ErrAPIKeyNotFound
	}

	var info APIKeyInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return nil, &APIKeyError{Type: "parse_error", Message: "failed to parse API key data", Cause: err}
	}

	// Check if key is enabled
	if !info.Enabled {
		return nil, ErrAPIKeyDisabled
	}

	// Check if key has expired
	if !info.ExpiresAt.IsZero() && time.Now().After(info.ExpiresAt) {
		return nil, ErrAPIKeyExpired
	}

	// Update last used timestamp
	info.LastUsed = time.Now()
	if err := a.updateAPIKeyInfo(ctx, apiKey, info); err != nil {
		// Log error but don't fail validation
		// In production, you might want to use a logger here
	}

	return &info, nil
}

// CreateAPIKey creates a new API key for a user
func (a *APIKeyManager) CreateAPIKey(ctx context.Context, userID string, rateLimit int, description string, expiresIn time.Duration) (string, error) {
	if userID == "" {
		return "", ErrInvalidUserID
	}

	if rateLimit < 0 {
		return "", ErrInvalidRateLimit
	}

	// Generate a secure API key
	apiKey := generateSecureAPIKey()

	// Check if key already exists (very unlikely but possible)
	key := "apikey:" + apiKey
	exists, err := a.storage.Exists(ctx, key)
	if err != nil {
		return "", &APIKeyError{Type: "storage_error", Message: "failed to check key existence", Cause: err}
	}
	if exists {
		return "", ErrAPIKeyExists
	}

	// Create API key info
	now := time.Now()
	info := APIKeyInfo{
		UserID:      userID,
		RateLimit:   rateLimit,
		Enabled:     true,
		CreatedAt:   now,
		Description: description,
	}

	// Set expiration if provided
	if expiresIn > 0 {
		info.ExpiresAt = now.Add(expiresIn)
	}

	// Store API key info
	if err := a.storeAPIKeyInfo(ctx, apiKey, info); err != nil {
		return "", err
	}

	// Add to user's key list
	if err := a.addToUserKeyList(ctx, userID, apiKey); err != nil {
		// Clean up the created key if we can't add to user list
		a.storage.Delete(ctx, key)
		return "", err
	}

	return apiKey, nil
}

// RevokeAPIKey disables an API key
func (a *APIKeyManager) RevokeAPIKey(ctx context.Context, apiKey string) error {
	if apiKey == "" {
		return ErrInvalidAPIKey
	}

	key := "apikey:" + apiKey
	data, err := a.storage.Get(ctx, key)
	if err != nil {
		return ErrAPIKeyNotFound
	}

	var info APIKeyInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return &APIKeyError{Type: "parse_error", Message: "failed to parse API key data", Cause: err}
	}

	info.Enabled = false
	return a.updateAPIKeyInfo(ctx, apiKey, info)
}

// ListAPIKeys returns all API keys for a user
func (a *APIKeyManager) ListAPIKeys(ctx context.Context, userID string) (*APIKeyList, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}

	// Get user's key list
	userKey := "userkeys:" + userID
	data, err := a.storage.Get(ctx, userKey)
	if err != nil {
		// Return empty list if user has no keys
		return &APIKeyList{UserID: userID, Keys: []APIKeyInfo{}}, nil
	}

	var keyList []string
	if err := json.Unmarshal([]byte(data), &keyList); err != nil {
		return nil, &APIKeyError{Type: "parse_error", Message: "failed to parse user key list", Cause: err}
	}

	// Get info for each key
	var keys []APIKeyInfo
	for _, apiKey := range keyList {
		key := "apikey:" + apiKey
		data, err := a.storage.Get(ctx, key)
		if err != nil {
			// Skip keys that can't be retrieved
			continue
		}

		var info APIKeyInfo
		if err := json.Unmarshal([]byte(data), &info); err != nil {
			// Skip keys with invalid data
			continue
		}

		keys = append(keys, info)
	}

	return &APIKeyList{UserID: userID, Keys: keys}, nil
}

// UpdateAPIKey updates API key properties
func (a *APIKeyManager) UpdateAPIKey(ctx context.Context, apiKey string, updates map[string]interface{}) error {
	if apiKey == "" {
		return ErrInvalidAPIKey
	}

	key := "apikey:" + apiKey
	data, err := a.storage.Get(ctx, key)
	if err != nil {
		return ErrAPIKeyNotFound
	}

	var info APIKeyInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return &APIKeyError{Type: "parse_error", Message: "failed to parse API key data", Cause: err}
	}

	// Apply updates
	for field, value := range updates {
		switch field {
		case "rate_limit":
			if rateLimit, ok := value.(int); ok && rateLimit >= 0 {
				info.RateLimit = rateLimit
			} else {
				return ErrInvalidRateLimit
			}
		case "enabled":
			if enabled, ok := value.(bool); ok {
				info.Enabled = enabled
			}
		case "description":
			if desc, ok := value.(string); ok {
				info.Description = desc
			}
		case "expires_at":
			if expiresAt, ok := value.(time.Time); ok {
				info.ExpiresAt = expiresAt
			}
		}
	}

	return a.updateAPIKeyInfo(ctx, apiKey, info)
}

// DeleteAPIKey permanently removes an API key
func (a *APIKeyManager) DeleteAPIKey(ctx context.Context, apiKey string) error {
	if apiKey == "" {
		return ErrInvalidAPIKey
	}

	key := "apikey:" + apiKey
	data, err := a.storage.Get(ctx, key)
	if err != nil {
		return ErrAPIKeyNotFound
	}

	var info APIKeyInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return &APIKeyError{Type: "parse_error", Message: "failed to parse API key data", Cause: err}
	}

	// Remove from user's key list
	if err := a.removeFromUserKeyList(ctx, info.UserID, apiKey); err != nil {
		return &APIKeyError{Type: "cleanup_error", Message: "failed to remove from user key list", Cause: err}
	}

	// Delete the API key
	return a.storage.Delete(ctx, key)
}

// Helper methods

func generateSecureAPIKey() string {
	// Generate 32 random bytes (256 bits)
	bytes := make([]byte, 32)
	rand.Read(bytes)

	// Convert to hex and add prefix
	hexKey := hex.EncodeToString(bytes)
	return "kalshi_" + hexKey
}

func (a *APIKeyManager) storeAPIKeyInfo(ctx context.Context, apiKey string, info APIKeyInfo) error {
	key := "apikey:" + apiKey
	data, err := json.Marshal(info)
	if err != nil {
		return &APIKeyError{Type: "serialization_error", Message: "failed to serialize API key info", Cause: err}
	}

	// Set TTL based on expiration
	var ttl time.Duration
	if !info.ExpiresAt.IsZero() {
		ttl = time.Until(info.ExpiresAt)
		if ttl <= 0 {
			return ErrAPIKeyExpired
		}
	} else {
		// Default TTL of 1 year for keys without expiration
		ttl = 365 * 24 * time.Hour
	}

	return a.storage.Set(ctx, key, string(data), ttl)
}

func (a *APIKeyManager) updateAPIKeyInfo(ctx context.Context, apiKey string, info APIKeyInfo) error {
	key := "apikey:" + apiKey
	data, err := json.Marshal(info)
	if err != nil {
		return &APIKeyError{Type: "serialization_error", Message: "failed to serialize API key info", Cause: err}
	}

	// Calculate TTL based on expiration
	var ttl time.Duration
	if !info.ExpiresAt.IsZero() {
		ttl = time.Until(info.ExpiresAt)
		if ttl <= 0 {
			return ErrAPIKeyExpired
		}
	} else {
		// Default TTL of 1 year for keys without expiration
		ttl = 365 * 24 * time.Hour
	}

	return a.storage.Set(ctx, key, string(data), ttl)
}

func (a *APIKeyManager) addToUserKeyList(ctx context.Context, userID, apiKey string) error {
	userKey := "userkeys:" + userID
	var keyList []string

	// Try to get existing list
	data, err := a.storage.Get(ctx, userKey)
	if err == nil {
		if err := json.Unmarshal([]byte(data), &keyList); err != nil {
			return &APIKeyError{Type: "parse_error", Message: "failed to parse user key list", Cause: err}
		}
	}

	// Add new key
	keyList = append(keyList, apiKey)

	// Store updated list
	dataBytes, err := json.Marshal(keyList)
	if err != nil {
		return &APIKeyError{Type: "serialization_error", Message: "failed to serialize user key list", Cause: err}
	}
	data = string(dataBytes)

	// Store with long TTL (same as API keys)
	return a.storage.Set(ctx, userKey, string(data), 365*24*time.Hour)
}

func (a *APIKeyManager) removeFromUserKeyList(ctx context.Context, userID, apiKey string) error {
	userKey := "userkeys:" + userID
	data, err := a.storage.Get(ctx, userKey)
	if err != nil {
		return nil // User has no keys, nothing to remove
	}

	var keyList []string
	if err := json.Unmarshal([]byte(data), &keyList); err != nil {
		return &APIKeyError{Type: "parse_error", Message: "failed to parse user key list", Cause: err}
	}

	// Remove the key
	var newList []string
	for _, key := range keyList {
		if key != apiKey {
			newList = append(newList, key)
		}
	}

	// Store updated list
	dataBytes, err := json.Marshal(newList)
	if err != nil {
		return &APIKeyError{Type: "serialization_error", Message: "failed to serialize user key list", Cause: err}
	}
	data = string(dataBytes)

	return a.storage.Set(ctx, userKey, string(data), 365*24*time.Hour)
}
