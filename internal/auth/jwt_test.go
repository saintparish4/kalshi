package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTManager(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-for-security"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	manager := NewJWTManager(secret, accessExpiry, refreshExpiry)
	assert.NotNil(t, manager)
	assert.Equal(t, secret, manager.secret)
	assert.Equal(t, accessExpiry, manager.accessExpiry)
	assert.Equal(t, refreshExpiry, manager.refreshExpiry)
	assert.NotNil(t, manager.logger)
}

func TestGenerateTokenPair(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-for-security"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	manager := NewJWTManager(secret, accessExpiry, refreshExpiry)

	userID := "user123"
	role := "admin"

	tokenPair, err := manager.GenerateTokenPair(userID, role)
	require.NoError(t, err)
	assert.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Greater(t, tokenPair.ExpiresIn, time.Now().Unix())
}

func TestValidateToken_Success(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-for-security"
	manager := NewJWTManager(secret, 15*time.Minute, 7*24*time.Hour)

	userID := "user123"
	role := "admin"

	tokenPair, err := manager.GenerateTokenPair(userID, role)
	require.NoError(t, err)

	ctx := context.Background()
	claims, err := manager.ValidateToken(ctx, tokenPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, "access", claims.Type)
}

func TestValidateToken_Expired(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-for-security"
	// Use very short expiry for testing
	manager := NewJWTManager(secret, 1*time.Millisecond, 7*24*time.Hour)

	userID := "user123"
	role := "admin"

	tokenPair, err := manager.GenerateTokenPair(userID, role)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	ctx := context.Background()
	_, err = manager.ValidateToken(ctx, tokenPair.AccessToken)
	assert.Error(t, err)

	var jwtErr *JWTError
	assert.ErrorAs(t, err, &jwtErr)
	assert.Equal(t, "token_expired", jwtErr.Type)
}

func TestValidateToken_InvalidAlgorithm(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-for-security"
	manager := NewJWTManager(secret, 15*time.Minute, 7*24*time.Hour)

	// Create a token with RS256 algorithm (different from HS256)
	claims := Claims{
		UserID: "user123",
		Role:   "admin",
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Generate a real RSA private key for signing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = manager.ValidateToken(ctx, tokenString)
	assert.Error(t, err)

	var jwtErr *JWTError
	assert.ErrorAs(t, err, &jwtErr)
	assert.Equal(t, "invalid_algorithm", jwtErr.Type)
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-for-security"
	manager := NewJWTManager(secret, 15*time.Minute, 7*24*time.Hour)

	// Create a token with different secret
	wrongSecret := "wrong-secret-key-that-is-long-enough-for-security"
	claims := Claims{
		UserID: "user123",
		Role:   "admin",
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(wrongSecret))
	require.NoError(t, err)

	ctx := context.Background()
	_, err = manager.ValidateToken(ctx, tokenString)
	assert.Error(t, err)

	var jwtErr *JWTError
	assert.ErrorAs(t, err, &jwtErr)
	assert.Equal(t, "parsing_failed", jwtErr.Type)
}

func TestValidateToken_NotValidYet(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-for-security"
	manager := NewJWTManager(secret, 15*time.Minute, 7*24*time.Hour)

	// Create a token that's not valid yet
	claims := Claims{
		UserID: "user123",
		Role:   "admin",
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // Not valid for 1 hour
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	ctx := context.Background()
	_, err = manager.ValidateToken(ctx, tokenString)
	assert.Error(t, err)

	var jwtErr *JWTError
	assert.ErrorAs(t, err, &jwtErr)
	assert.Equal(t, "token_not_valid_yet", jwtErr.Type)
}

func TestValidateAccessToken_Success(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-for-security"
	manager := NewJWTManager(secret, 15*time.Minute, 7*24*time.Hour)

	userID := "user123"
	role := "admin"

	tokenPair, err := manager.GenerateTokenPair(userID, role)
	require.NoError(t, err)

	ctx := context.Background()
	claims, err := manager.ValidateAccessToken(ctx, tokenPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, "access", claims.Type)
}

func TestValidateAccessToken_RefreshToken(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-for-security"
	manager := NewJWTManager(secret, 15*time.Minute, 7*24*time.Hour)

	userID := "user123"
	role := "admin"

	tokenPair, err := manager.GenerateTokenPair(userID, role)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = manager.ValidateAccessToken(ctx, tokenPair.RefreshToken)
	assert.Error(t, err)

	var jwtErr *JWTError
	assert.ErrorAs(t, err, &jwtErr)
	assert.Equal(t, "invalid_token_type", jwtErr.Type)
}

func TestRefreshToken_Success(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-for-security"
	manager := NewJWTManager(secret, 15*time.Minute, 7*24*time.Hour)

	userID := "user123"
	role := "admin"

	tokenPair, err := manager.GenerateTokenPair(userID, role)
	require.NoError(t, err)

	ctx := context.Background()
	newTokenPair, err := manager.RefreshToken(ctx, tokenPair.RefreshToken)
	require.NoError(t, err)
	assert.NotNil(t, newTokenPair)
	assert.NotEmpty(t, newTokenPair.AccessToken)
	assert.NotEmpty(t, newTokenPair.RefreshToken)
	assert.NotEqual(t, tokenPair.AccessToken, newTokenPair.AccessToken)
	assert.NotEqual(t, tokenPair.RefreshToken, newTokenPair.RefreshToken)

	// Verify the new access token is valid
	claims, err := manager.ValidateAccessToken(ctx, newTokenPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, role, claims.Role)
}

func TestRefreshToken_AccessToken(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-for-security"
	manager := NewJWTManager(secret, 15*time.Minute, 7*24*time.Hour)

	userID := "user123"
	role := "admin"

	tokenPair, err := manager.GenerateTokenPair(userID, role)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = manager.RefreshToken(ctx, tokenPair.AccessToken)
	assert.Error(t, err)

	var jwtErr *JWTError
	assert.ErrorAs(t, err, &jwtErr)
	assert.Equal(t, "invalid_token_type", jwtErr.Type)
}

func TestRefreshToken_ExpiredRefreshToken(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-for-security"
	// Use very short expiry for refresh token
	manager := NewJWTManager(secret, 15*time.Minute, 1*time.Millisecond)

	userID := "user123"
	role := "admin"

	tokenPair, err := manager.GenerateTokenPair(userID, role)
	require.NoError(t, err)

	// Wait for refresh token to expire
	time.Sleep(10 * time.Millisecond)

	ctx := context.Background()
	_, err = manager.RefreshToken(ctx, tokenPair.RefreshToken)
	assert.Error(t, err)

	var jwtErr *JWTError
	assert.ErrorAs(t, err, &jwtErr)
	assert.Equal(t, "token_expired", jwtErr.Type)
}

func TestJWTError_Error(t *testing.T) {
	originalErr := jwt.ErrTokenExpired
	jwtErr := &JWTError{
		Type:    "test_error",
		Message: "test message",
		Cause:   originalErr,
	}

	errorStr := jwtErr.Error()
	assert.Contains(t, errorStr, "JWT test_error")
	assert.Contains(t, errorStr, "test message")
	assert.Contains(t, errorStr, "token is expired")
}

func TestJWTError_Unwrap(t *testing.T) {
	originalErr := jwt.ErrTokenExpired
	jwtErr := &JWTError{
		Type:    "test_error",
		Message: "test message",
		Cause:   originalErr,
	}

	unwrapped := jwtErr.Unwrap()
	assert.Equal(t, originalErr, unwrapped)
}

func TestJWTError_WithoutCause(t *testing.T) {
	jwtErr := &JWTError{
		Type:    "test_error",
		Message: "test message",
	}

	errorStr := jwtErr.Error()
	assert.Contains(t, errorStr, "JWT test_error")
	assert.Contains(t, errorStr, "test message")
	assert.NotContains(t, errorStr, "cause:")
}
