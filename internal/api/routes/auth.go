package routes

import (
	"kalshi/internal/api/handlers"
	"kalshi/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// setupAuthRoutes configures authentication endpoints
func setupAuthRoutes(router *gin.Engine, cfg *RouterConfig) {
	authHandler := handlers.NewAuthHandler(cfg.JWTManager, cfg.APIKeyManager, cfg.Logger)

	auth := router.Group("/auth")

	// Public authentication endpoints
	public := auth.Group("/")
	{
		public.POST("/login", authHandler.Login)
		public.POST("/register", authHandler.Register)
		public.POST("/refresh", authHandler.RefreshToken)
		public.POST("/forgot-password", authHandler.ForgotPassword)
		public.POST("/reset-password", authHandler.ResetPassword)
	}

	// Protected authentication endpoints
	protected := auth.Group("/")
	protected.Use(middleware.JWTAuth(cfg.JWTManager, cfg.Logger))
	{
		protected.POST("/logout", authHandler.Logout)
		protected.GET("/profile", authHandler.GetProfile)
		protected.PUT("/profile", authHandler.UpdateProfile)
		protected.POST("/change-password", authHandler.ChangePassword)
		protected.GET("/sessions", authHandler.GetSessions)
		protected.DELETE("/sessions/:id", authHandler.RevokeSession)
	}

	// API Key management
	apiKeys := auth.Group("/apikeys")
	apiKeys.Use(middleware.JWTAuth(cfg.JWTManager, cfg.Logger))
	{
		apiKeys.GET("", authHandler.GetAPIKeys)
		apiKeys.POST("", authHandler.CreateAPIKey)
		apiKeys.DELETE("/:id", authHandler.RevokeAPIKey)
		apiKeys.PUT("/:id/regenerate", authHandler.RegenerateAPIKey)
	}

	// OAuth endpoints (if implemented)
	oauth := auth.Group("/oauth")
	{
		oauth.GET("/:provider", authHandler.OAuthLogin)
		oauth.GET("/:provider/callback", authHandler.OAuthCallback)
	}

	// Token validation (for other services)
	token := auth.Group("/token")
	{
		token.POST("/validate", authHandler.ValidateToken)
		token.POST("/introspect", authHandler.IntrospectToken)
	}
}

// Auth handler methods would be implemented in handlers/auth.go
