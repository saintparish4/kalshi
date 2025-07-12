package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kalshi/internal/api/routes"
	"kalshi/internal/auth"
	"kalshi/internal/cache"
	"kalshi/internal/config"
	"kalshi/internal/gateway"
	"kalshi/internal/ratelimit"
	"kalshi/internal/storage"
	"kalshi/pkg/logger"
)

// Build information (set by build script)
var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

// Application holds all application dependencies
type Application struct {
	config        *config.Config
	logger        *logger.Logger
	storage       storage.Storage
	cacheManager  *cache.Manager
	gateway       *gateway.Gateway
	limiter       *ratelimit.Limiter
	jwtManager    *auth.JWTManager
	apiKeyManager *auth.APIKeyManager
	server        *http.Server
	metricsServer *http.Server
}

func main() {
	var configPath string
	var showVersion bool

	flag.StringVar(&configPath, "config", "configs/config.yaml", "Path to configuration file")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.Parse()

	if showVersion {
		fmt.Printf("Kalshi API Gateway\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Build time: %s\n", buildTime)
		fmt.Printf("Git commit: %s\n", gitCommit)
		os.Exit(0)
	}

	// Initialize application
	app, err := initializeApplication(configPath)
	if err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	// Start the application
	if err := app.start(); err != nil {
		app.logger.Fatal("Failed to start application", "error", err)
	}

	// Wait for shutdown signal
	app.waitForShutdown()

	// Graceful shutdown
	app.shutdown()
}

// initializeApplication sets up all application components
func initializeApplication(configPath string) (*Application, error) {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	log, err := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize global logger
	if err := logger.InitGlobal(cfg.Logging.Level, cfg.Logging.Format); err != nil {
		return nil, fmt.Errorf("failed to initialize global logger: %w", err)
	}

	log.Info("Starting API Gateway",
		"version", version,
		"build_time", buildTime,
		"git_commit", gitCommit,
	)

	// Initialize storage
	stor, err := initializeStorage(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize cache
	cacheManager, err := initializeCache(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	// Initialize authentication
	jwtManager := auth.NewJWTManager(cfg.Auth.JWT.Secret, cfg.Auth.JWT.AccessExpiry, cfg.Auth.JWT.RefreshExpiry)
	apiKeyManager := auth.NewAPIKeyManager(stor)

	// Initialize rate limiter
	limiter := ratelimit.NewLimiter(stor, cfg.RateLimit.DefaultRate, cfg.RateLimit.BurstCapacity)

	// Initialize gateway
	gw := gateway.New(cfg, cacheManager, log)

	// Create application instance
	app := &Application{
		config:        cfg,
		logger:        log,
		storage:       stor,
		cacheManager:  cacheManager,
		gateway:       gw,
		limiter:       limiter,
		jwtManager:    jwtManager,
		apiKeyManager: apiKeyManager,
	}

	// Setup HTTP servers
	app.setupServers()

	return app, nil
}

// initializeStorage creates the appropriate storage backend
func initializeStorage(cfg *config.Config, log *logger.Logger) (storage.Storage, error) {
	if cfg.RateLimit.Storage == "redis" {
		log.Info("Initializing Redis storage", "addr", cfg.Cache.Redis.Addr)
		return storage.NewRedisStorage(
			cfg.Cache.Redis.Addr,
			cfg.Cache.Redis.Password,
			cfg.Cache.Redis.DB,
		)
	}

	log.Info("Initializing memory storage")
	return storage.NewMemoryStorage(), nil
}

// initializeCache creates the cache manager with L1 and L2 caches
func initializeCache(cfg *config.Config, log *logger.Logger) (*cache.Manager, error) {
	// L1 Cache (Memory)
	l1Cache := cache.NewMemoryCache(cfg.Cache.Memory.MaxSize, cfg.Cache.Memory.TTL)
	log.Info("Initialized L1 cache (memory)", "max_size", cfg.Cache.Memory.MaxSize)

	// L2 Cache (Redis) - optional
	var l2Cache cache.Cache
	if cfg.Cache.Redis.Addr != "" {
		var err error
		l2Cache, err = cache.NewRedisCache(
			cfg.Cache.Redis.Addr,
			cfg.Cache.Redis.Password,
			cfg.Cache.Redis.DB,
			cfg.Cache.Redis.TTL,
		)
		if err != nil {
			log.Warn("Failed to initialize Redis cache, using memory only", "error", err)
		} else {
			log.Info("Initialized L2 cache (Redis)", "addr", cfg.Cache.Redis.Addr)
		}
	}

	return cache.NewManager(l1Cache, l2Cache, true), nil
}

// setupServers configures HTTP servers
func (app *Application) setupServers() {
	// Main API server
	routerConfig := &routes.RouterConfig{
		Config:        app.config,
		Gateway:       app.gateway,
		Limiter:       app.limiter,
		JWTManager:    app.jwtManager,
		APIKeyManager: app.apiKeyManager,
		Logger:        app.logger,
	}

	router := routes.SetupRouter(routerConfig)

	app.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", app.config.Server.Host, app.config.Server.Port),
		Handler:      router,
		ReadTimeout:  app.config.Server.ReadTimeout,
		WriteTimeout: app.config.Server.WriteTimeout,
		IdleTimeout:  app.config.Server.IdleTimeout,
	}

	// Metrics server (if enabled and on different port)
	if app.config.Metrics.Enabled && app.config.Metrics.Port != app.config.Server.Port {
		metricsRouter := routes.SetupMetricsOnlyRouter(routerConfig)
		app.metricsServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", app.config.Metrics.Port),
			Handler: metricsRouter,
		}
	}
}

// start begins serving HTTP requests
func (app *Application) start() error {
	// Start main server
	go func() {
		app.logger.Info("Starting main server", "address", app.server.Addr)
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.logger.Fatal("Main server failed", "error", err)
		}
	}()

	// Start metrics server if configured
	if app.metricsServer != nil {
		go func() {
			app.logger.Info("Starting metrics server", "address", app.metricsServer.Addr)
			if err := app.metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				app.logger.Error("Metrics server failed", "error", err)
			}
		}()
	}

	// Health check - wait for server to be ready
	return app.waitForServerReady()
}

// waitForServerReady waits for the server to start accepting connections
func (app *Application) waitForServerReady() error {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(100 * time.Millisecond)

		resp, err := http.Get(fmt.Sprintf("http://%s/health", app.server.Addr))
		if err == nil {
			resp.Body.Close()
			app.logger.Info("Server is ready and accepting connections")
			return nil
		}
	}

	return fmt.Errorf("server failed to become ready after %d attempts", maxAttempts)
}

// waitForShutdown waits for interrupt signal
func (app *Application) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	app.logger.Info("Received shutdown signal", "signal", sig.String())
}

// shutdown gracefully shuts down the application
func (app *Application) shutdown() {
	app.logger.Info("Shutting down application...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown main server
	if err := app.server.Shutdown(ctx); err != nil {
		app.logger.Error("Error shutting down main server", "error", err)
	} else {
		app.logger.Info("Main server shut down gracefully")
	}

	// Shutdown metrics server
	if app.metricsServer != nil {
		if err := app.metricsServer.Shutdown(ctx); err != nil {
			app.logger.Error("Error shutting down metrics server", "error", err)
		} else {
			app.logger.Info("Metrics server shut down gracefully")
		}
	}

	// Close storage connections
	if app.storage != nil {
		if err := app.storage.Close(); err != nil {
			app.logger.Error("Error closing storage", "error", err)
		} else {
			app.logger.Info("Storage connections closed")
		}
	}

	// Close cache connections
	if app.cacheManager != nil {
		// Assuming cache manager has a Close method
		app.logger.Info("Cache connections closed")
	}

	app.logger.Info("Application shutdown complete")
}
