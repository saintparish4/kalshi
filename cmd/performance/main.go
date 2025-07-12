package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	// Parse command line flags
	var (
		baseURL        = flag.String("base-url", "http://localhost:8080", "Base URL for the API Gateway")
		duration       = flag.Duration("duration", 1*time.Minute, "Test duration")
		initialRPS     = flag.Int("initial-rps", 50, "Initial requests per second")
		maxRPS         = flag.Int("max-rps", 500, "Maximum requests per second")
		rampUpDuration = flag.Duration("ramp-up", 30*time.Second, "Ramp-up duration")
		warmupDuration = flag.Duration("warmup", 10*time.Second, "Warmup duration")
		maxConcurrency = flag.Int("concurrency", 100, "Maximum concurrent connections")
		workerPoolSize = flag.Int("workers", 50, "Worker pool size")
		outputFile     = flag.String("output", "", "Output file for results (optional)")
		useTimestamp   = flag.Bool("timestamp", false, "Use timestamp in filename (default: false)")
		verbose        = flag.Bool("verbose", false, "Enable verbose output")

		// Authentication flags
		useAuth      = flag.Bool("auth", false, "Enable authentication for tests")
		authType     = flag.String("auth-type", "jwt", "Authentication type: jwt, api-key, or both")
		jwtSecret    = flag.String("jwt-secret", "your-secret-key-change-this-in-production-must-be-32-chars", "JWT secret for token generation")
		apiKey       = flag.String("api-key", "", "API key for authentication")
		apiKeyHeader = flag.String("api-key-header", "X-API-Key", "API key header name")
		username     = flag.String("username", "test-user", "Username for JWT token generation")
		userRole     = flag.String("role", "user", "User role for JWT token generation")
	)
	flag.Parse()

	// Override with environment variables if provided
	if envBaseURL := os.Getenv("TEST_BASE_URL"); envBaseURL != "" {
		*baseURL = envBaseURL
	}
	if envMaxRPS := os.Getenv("TEST_MAX_RPS"); envMaxRPS != "" {
		fmt.Sscanf(envMaxRPS, "%d", maxRPS)
	}
	if envWorkerPool := os.Getenv("TEST_WORKER_POOL"); envWorkerPool != "" {
		fmt.Sscanf(envWorkerPool, "%d", workerPoolSize)
	}
	if envAPIKey := os.Getenv("TEST_API_KEY"); envAPIKey != "" {
		*apiKey = envAPIKey
	}
	if envJWTSecret := os.Getenv("TEST_JWT_SECRET"); envJWTSecret != "" {
		*jwtSecret = envJWTSecret
	}

	config := TestConfig{
		BaseURL:        *baseURL,
		Duration:       *duration,
		InitialRPS:     *initialRPS,
		MaxRPS:         *maxRPS,
		RampUpDuration: *rampUpDuration,
		WarmupDuration: *warmupDuration,
		MaxConcurrency: *maxConcurrency,
		WorkerPoolSize: *workerPoolSize,
		UseAuth:        *useAuth,
		AuthType:       *authType,
		JWTSecret:      *jwtSecret,
		APIKey:         *apiKey,
		APIKeyHeader:   *apiKeyHeader,
		Username:       *username,
		UserRole:       *userRole,
		TestEndpoints: []string{
			"/api/v1/health",
			"/api/v1/metrics",
			"/api/v1/proxy/backend1/users",
			"/api/v1/proxy/backend1/posts",
		},
	}

	if *verbose {
		fmt.Printf("Performance Test Configuration:\n")
		fmt.Printf("  Base URL: %s\n", config.BaseURL)
		fmt.Printf("  Duration: %v\n", config.Duration)
		fmt.Printf("  Initial RPS: %d\n", config.InitialRPS)
		fmt.Printf("  Max RPS: %d\n", config.MaxRPS)
		fmt.Printf("  Ramp-up Duration: %v\n", config.RampUpDuration)
		fmt.Printf("  Warmup Duration: %v\n", config.WarmupDuration)
		fmt.Printf("  Max Concurrency: %d\n", config.MaxConcurrency)
		fmt.Printf("  Worker Pool Size: %d\n", config.WorkerPoolSize)
		fmt.Printf("  Authentication: %v\n", config.UseAuth)
		if config.UseAuth {
			fmt.Printf("  Auth Type: %s\n", config.AuthType)
			fmt.Printf("  Username: %s\n", config.Username)
			fmt.Printf("  User Role: %s\n", config.UserRole)
			if config.AuthType == "api-key" || config.AuthType == "both" {
				fmt.Printf("  API Key Header: %s\n", config.APIKeyHeader)
			}
		}
		fmt.Printf("  Test Endpoints: %v\n", config.TestEndpoints)
		fmt.Println()
	}

	test := NewPerformanceTest(config)
	result := test.runRampUpTest()

	test.printResults(result)

	// Save results to JSON file if output file is specified
	if *outputFile != "" {
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Printf("Error marshaling results: %v", err)
			os.Exit(1)
		}

		if err := os.WriteFile(*outputFile, jsonData, 0644); err != nil {
			log.Printf("Error writing results to file: %v", err)
			os.Exit(1)
		}
		fmt.Printf("Results saved to %s\n", *outputFile)
	} else {
		// Default behavior: save with fixed filename or timestamp if requested
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Printf("Error marshaling results: %v", err)
			os.Exit(1)
		}

		var filename string
		if *useTimestamp {
			filename = fmt.Sprintf("performance_results_%s.json", time.Now().Format("20060102_150405"))
		} else {
			filename = "performance_results.json"
		}

		if err := os.WriteFile(filename, jsonData, 0644); err != nil {
			log.Printf("Error writing results to file: %v", err)
			os.Exit(1)
		} else {
			fmt.Printf("Results saved to %s\n", filename)
		}
	}
}
