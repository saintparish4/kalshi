package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kalshi/internal/auth"
	"kalshi/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock storage for testing
type MockStorage struct {
	data map[string]string
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string]string),
	}
}

func (m *MockStorage) Get(ctx context.Context, key string) (string, error) {
	if value, exists := m.data[key]; exists {
		return value, nil
	}
	return "", assert.AnError
}

func (m *MockStorage) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *MockStorage) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.data[key]
	return exists, nil
}

func (m *MockStorage) Increment(ctx context.Context, key string, by int64) (int64, error) {
	// Simple increment implementation for testing
	return 1, nil
}

func (m *MockStorage) Close() error {
	return nil
}

func TestJWTAuth(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	// Create a real JWT manager for testing
	jwtManager := auth.NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Authorization header required"}`,
		},
		{
			name:           "Invalid auth format - no Bearer",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid authorization header format. Use: Bearer <token>"}`,
		},
		{
			name:           "Invalid auth format - wrong parts",
			authHeader:     "Bearer token extra",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid authorization header format. Use: Bearer <token>"}`,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid or expired token"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(JWTAuth(jwtManager, log))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestAPIKeyAuth(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	// Create a real API key manager with mock storage
	mockStorage := NewMockStorage()
	apiKeyManager := auth.NewAPIKeyManager(mockStorage)

	tests := []struct {
		name           string
		headerKey      string
		headerValue    string
		queryKey       string
		queryValue     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Missing API key in header and query",
			headerKey:      "X-API-Key",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"API key required in header: X-API-Key"}`,
		},
		{
			name:           "Invalid API key",
			headerKey:      "X-API-Key",
			headerValue:    "invalid-key",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid API key"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(APIKeyAuth(apiKeyManager, tt.headerKey, log))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.headerValue != "" {
				req.Header.Set(tt.headerKey, tt.headerValue)
			}
			if tt.queryValue != "" {
				q := req.URL.Query()
				q.Set(tt.queryKey, tt.queryValue)
				req.URL.RawQuery = q.Encode()
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestOptionalAuth(t *testing.T) {
	log, _ := logger.New("info", "console")
	gin.SetMode(gin.TestMode)

	// Create real managers for testing
	jwtManager := auth.NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour)
	mockStorage := NewMockStorage()
	apiKeyManager := auth.NewAPIKeyManager(mockStorage)

	tests := []struct {
		name           string
		authHeader     string
		apiKeyHeader   string
		apiKeyValue    string
		queryKey       string
		queryValue     string
		expectedStatus int
		expectedUserID string
		expectedMethod string
	}{
		{
			name:           "No authentication - anonymous",
			expectedStatus: http.StatusOK,
			expectedUserID: AnonymousUser,
			expectedMethod: AuthMethodNone,
		},
		{
			name:           "Invalid JWT and API key - anonymous",
			authHeader:     "Bearer invalid-jwt",
			apiKeyHeader:   "X-API-Key",
			apiKeyValue:    "invalid-api-key",
			expectedStatus: http.StatusOK,
			expectedUserID: AnonymousUser,
			expectedMethod: AuthMethodNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(OptionalAuth(jwtManager, apiKeyManager, tt.apiKeyHeader, log))
			router.GET("/test", func(c *gin.Context) {
				userID, _ := c.Get(ContextUserID)
				authMethod, _ := c.Get(ContextAuthMethod)
				c.JSON(http.StatusOK, gin.H{
					"user_id":     userID,
					"auth_method": authMethod,
				})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.apiKeyValue != "" {
				req.Header.Set(tt.apiKeyHeader, tt.apiKeyValue)
			}
			if tt.queryValue != "" {
				q := req.URL.Query()
				q.Set(tt.queryKey, tt.queryValue)
				req.URL.RawQuery = q.Encode()
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				expectedBody := gin.H{
					"user_id":     tt.expectedUserID,
					"auth_method": tt.expectedMethod,
				}
				assert.JSONEq(t, toJSON(expectedBody), w.Body.String())
			}
		})
	}
}

// Helper function to convert gin.H to JSON string
func toJSON(data gin.H) string {
	// This is a simplified version - in a real test you'd use a proper JSON encoder
	// For now, we'll just return a basic string representation
	return `{"auth_method":"` + data["auth_method"].(string) + `","user_id":"` + data["user_id"].(string) + `"}`
}
