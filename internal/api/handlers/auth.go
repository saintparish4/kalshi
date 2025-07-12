package handlers

import (
	"net/http"
	"time"

	"kalshi/internal/auth"
	"kalshi/pkg/logger"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	jwtManager    *auth.JWTManager
	apiKeyManager *auth.APIKeyManager
	logger        *logger.Logger
}

func NewAuthHandler(jwtManager *auth.JWTManager, apiKeyManager *auth.APIKeyManager, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		jwtManager:    jwtManager,
		apiKeyManager: apiKeyManager,
		logger:        logger,
	}
}

// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var loginReq struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Basic authentication - validate credentials
	if loginReq.Username == "" || loginReq.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username and password are required",
		})
		return
	}

	// Simple validation (in production, check against database)
	if loginReq.Username == "admin" && loginReq.Password == "password" {
		tokenPair, err := h.jwtManager.GenerateTokenPair(loginReq.Username, "user")
		if err != nil {
			h.logger.Error("Failed to generate token", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Authentication failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token":      tokenPair.AccessToken,
			"type":       "Bearer",
			"expires_in": 15 * 60, // 15 minutes
		})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{
		"error": "Invalid credentials",
	})
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var registerReq struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&registerReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Basic registration validation
	if registerReq.Username == "" || registerReq.Email == "" || registerReq.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "All fields are required",
		})
		return
	}

	// Simple validation (in production, check for existing users)
	if registerReq.Username == "admin" {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Username already exists",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user_id": registerReq.Username,
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var refreshReq struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&refreshReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Basic token refresh validation
	if refreshReq.Token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Refresh token is required",
		})
		return
	}

	// Validate refresh token and generate new access token
	claims, err := h.jwtManager.ValidateToken(c.Request.Context(), refreshReq.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid refresh token",
		})
		return
	}

	// Generate new access token
	tokenPair, err := h.jwtManager.GenerateTokenPair(claims.UserID, claims.Type)
	if err != nil {
		h.logger.Error("Failed to generate new token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to refresh token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      tokenPair.AccessToken,
		"type":       "Bearer",
		"expires_in": 15 * 60,
	})
}

// ForgotPassword handles password reset requests
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var forgotReq struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&forgotReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email address",
		})
		return
	}

	// Basic password reset request validation
	if forgotReq.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email is required",
		})
		return
	}

	// In production, send actual reset email
	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var resetReq struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&resetReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Basic password reset validation
	if resetReq.Token == "" || resetReq.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token and new password are required",
		})
		return
	}

	// In production, validate reset token and update password
	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Basic logout - in production, blacklist the token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// In production, add token to blacklist
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
		"user_id": userID,
	})
}

// GetProfile returns user profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Basic profile retrieval
	c.JSON(http.StatusOK, gin.H{
		"user_id":    userID,
		"username":   userID,
		"email":      "user@example.com",
		"created_at": time.Now().UTC(),
		"updated_at": time.Now().UTC(),
	})
}

// UpdateProfile updates user profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var updateReq struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Basic profile update validation
	if updateReq.Email == "" && updateReq.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least one field must be provided",
		})
		return
	}

	// In production, update user profile in database
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user_id": userID,
	})
}

// ChangePassword changes user password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var changeReq struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&changeReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Basic password change validation
	if changeReq.CurrentPassword == "" || changeReq.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Current and new passwords are required",
		})
		return
	}

	// In production, validate current password and update
	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
		"user_id": userID,
	})
}

// GetSessions returns user sessions
func (h *AuthHandler) GetSessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Basic session retrieval
	c.JSON(http.StatusOK, gin.H{
		"sessions": []gin.H{
			{
				"id":         "session1",
				"user_id":    userID,
				"created_at": time.Now().UTC(),
				"last_used":  time.Now().UTC(),
			},
		},
	})
}

// RevokeSession revokes a specific session
func (h *AuthHandler) RevokeSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Session ID required",
		})
		return
	}

	// Basic session revocation
	c.JSON(http.StatusOK, gin.H{
		"message":    "Session revoked successfully",
		"user_id":    userID,
		"session_id": sessionID,
	})
}

// GetAPIKeys returns user API keys
func (h *AuthHandler) GetAPIKeys(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Basic API key retrieval
	c.JSON(http.StatusOK, gin.H{
		"api_keys": []gin.H{
			{
				"id":         "key1",
				"name":       "Default API Key",
				"created_at": time.Now().UTC(),
				"last_used":  time.Now().UTC(),
			},
		},
	})
}

// CreateAPIKey creates a new API key
func (h *AuthHandler) CreateAPIKey(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var createReq struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&createReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Basic API key creation
	if createReq.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "API key name is required",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"api_key": gin.H{
			"id":         "new_key_id",
			"name":       createReq.Name,
			"key":        "generated_api_key",
			"created_at": time.Now().UTC(),
		},
	})
}

// RevokeAPIKey revokes an API key
func (h *AuthHandler) RevokeAPIKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	keyID := c.Param("id")
	if keyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "API key ID required",
		})
		return
	}

	// Basic API key revocation
	c.JSON(http.StatusOK, gin.H{
		"message": "API key revoked successfully",
		"user_id": userID,
		"key_id":  keyID,
	})
}

// RegenerateAPIKey regenerates an API key
func (h *AuthHandler) RegenerateAPIKey(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	keyID := c.Param("id")
	if keyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "API key ID required",
		})
		return
	}

	// Basic API key regeneration
	c.JSON(http.StatusOK, gin.H{
		"api_key": gin.H{
			"id":         keyID,
			"name":       "Regenerated API Key",
			"key":        "new_generated_api_key",
			"created_at": time.Now().UTC(),
		},
	})
}

// OAuthLogin handles OAuth login
func (h *AuthHandler) OAuthLogin(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "OAuth provider required",
		})
		return
	}

	// Basic OAuth login
	c.JSON(http.StatusOK, gin.H{
		"message":  "OAuth login initiated",
		"provider": provider,
	})
}

// OAuthCallback handles OAuth callback
func (h *AuthHandler) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "OAuth provider required",
		})
		return
	}

	// Basic OAuth callback
	c.JSON(http.StatusOK, gin.H{
		"message":  "OAuth callback processed",
		"provider": provider,
	})
}

// ValidateToken validates a token
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var validateReq struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&validateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Basic token validation
	if validateReq.Token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token is required",
		})
		return
	}

	// In production, validate token with JWT manager
	c.JSON(http.StatusOK, gin.H{
		"valid":      true,
		"user_id":    "user123",
		"expires_at": time.Now().Add(15 * time.Minute).UTC(),
	})
}

// IntrospectToken introspects a token
func (h *AuthHandler) IntrospectToken(c *gin.Context) {
	var introspectReq struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&introspectReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Basic token introspection
	if introspectReq.Token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token is required",
		})
		return
	}

	// In production, introspect token with JWT manager
	c.JSON(http.StatusOK, gin.H{
		"active":  true,
		"user_id": "user123",
		"scope":   "read write",
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})
}
