package auth

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"kalshi/pkg/logger"

	"encoding/hex"

	"github.com/golang-jwt/jwt/v5"
)

// JWTError represents JWT-specific errors
type JWTError struct {
	Type    string
	Message string
	Cause   error
}

func (e *JWTError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("JWT %s: %s (cause: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("JWT %s: %s", e.Type, e.Message)
}

func (e *JWTError) Unwrap() error {
	return e.Cause
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secret        string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	logger        *logger.Logger
}

// Claims represents JWT token claims
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	Type   string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// NewJWTManager creates a new JWT manager instance
func NewJWTManager(secret string, accessExpiry, refreshExpiry time.Duration) *JWTManager {
	log, _ := logger.New("info", "json")
	return &JWTManager{
		secret:        secret,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
		logger:        log,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (j *JWTManager) GenerateTokenPair(userID, role string) (*TokenPair, error) {
	ctx := context.Background()

	// Generate access token
	accessToken, err := j.generateToken(ctx, userID, role, "access", j.accessExpiry)
	if err != nil {
		j.logger.Error("Failed to generate access token", "user_id", userID, "error", err)
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := j.generateToken(ctx, userID, role, "refresh", j.refreshExpiry)
	if err != nil {
		j.logger.Error("Failed to generate refresh token", "user_id", userID, "error", err)
		return nil, err
	}

	j.logger.Info("Generated token pair", "user_id", userID, "role", role)

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    time.Now().Add(j.accessExpiry).Unix(),
	}, nil
}

// generateToken generates a single token with specified type and expiry
func (j *JWTManager) generateToken(_ context.Context, userID, role, tokenType string, expiry time.Duration) (string, error) {
	now := time.Now()
	jti := make([]byte, 16)
	_, err := rand.Read(jti)
	if err != nil {
		return "", &JWTError{
			Type:    "generation_failed",
			Message: "failed to generate jti",
			Cause:   err,
		}
	}
	claims := Claims{
		UserID: userID,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        hex.EncodeToString(jti),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return "", &JWTError{
			Type:    "generation_failed",
			Message: "failed to sign token",
			Cause:   err,
		}
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	// First, parse the token without validation to check the algorithm
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, &JWTError{
			Type:    "parsing_failed",
			Message: "failed to parse token",
			Cause:   err,
		}
	}

	// Check algorithm before attempting to validate
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, &JWTError{
			Type:    "invalid_algorithm",
			Message: fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]),
		}
	}

	// Now parse with validation
	token, err = jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secret), nil
	})

	if err != nil {
		// Check for specific JWT validation errors
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, &JWTError{
				Type:    "token_expired",
				Message: "token has expired",
				Cause:   err,
			}
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, &JWTError{
				Type:    "token_not_valid_yet",
				Message: "token is not valid yet",
				Cause:   err,
			}
		}
		return nil, &JWTError{
			Type:    "parsing_failed",
			Message: "failed to parse token",
			Cause:   err,
		}
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, &JWTError{
			Type:    "invalid_token",
			Message: "token claims are invalid",
		}
	}

	j.logger.Debug("Token validated successfully", "user_id", claims.UserID, "type", claims.Type)
	return claims, nil
}

// RefreshToken generates a new access token using a valid refresh token
func (j *JWTManager) RefreshToken(ctx context.Context, refreshTokenString string) (*TokenPair, error) {
	// Validate the refresh token
	claims, err := j.ValidateToken(ctx, refreshTokenString)
	if err != nil {
		j.logger.Error("Failed to validate refresh token", "error", err)
		return nil, err
	}

	// Ensure this is actually a refresh token
	if claims.Type != "refresh" {
		return nil, &JWTError{
			Type:    "invalid_token_type",
			Message: "token is not a refresh token",
		}
	}

	// Generate new token pair
	tokenPair, err := j.GenerateTokenPair(claims.UserID, claims.Role)
	if err != nil {
		j.logger.Error("Failed to generate new token pair during refresh", "user_id", claims.UserID, "error", err)
		return nil, err
	}

	j.logger.Info("Token refreshed successfully", "user_id", claims.UserID)
	return tokenPair, nil
}

// ValidateAccessToken validates an access token specifically
func (j *JWTManager) ValidateAccessToken(ctx context.Context, tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	// Ensure this is an access token
	if claims.Type != "access" {
		return nil, &JWTError{
			Type:    "invalid_token_type",
			Message: "token is not an access token",
		}
	}

	return claims, nil
}
