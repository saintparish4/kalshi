package main

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestConfig holds the configuration for performance tests
type TestConfig struct {
	BaseURL        string        `json:"base_url"`
	Duration       time.Duration `json:"duration"`
	InitialRPS     int           `json:"initial_rps"`
	MaxRPS         int           `json:"max_rps"`
	RampUpDuration time.Duration `json:"ramp_up_duration"`
	NumUsers       int           `json:"num_users"`
	AuthToken      string        `json:"auth_token"`
	TestEndpoints  []string      `json:"test_endpoints"`
	WarmupDuration time.Duration `json:"warmup_duration"`
	MaxConcurrency int           `json:"max_concurrency"`
	WorkerPoolSize int           `json:"worker_pool_size"`
	AdminUsername  string        `json:"admin_username"`
	AdminPassword  string        `json:"admin_password"`
	
	// Authentication configuration
	UseAuth      bool   `json:"use_auth"`
	AuthType     string `json:"auth_type"`
	JWTSecret    string `json:"jwt_secret"`
	APIKey       string `json:"api_key"`
	APIKeyHeader string `json:"api_key_header"`
	Username     string `json:"username"`
	UserRole     string `json:"user_role"`
}

// TestResult holds the results of a performance test
type TestResult struct {
	TotalRequests   int64         `json:"total_requests"`
	SuccessfulReqs  int64         `json:"successful_requests"`
	FailedReqs      int64         `json:"failed_requests"`
	AvgLatency      time.Duration `json:"avg_latency"`
	P95Latency      time.Duration `json:"p95_latency"`
	P99Latency      time.Duration `json:"p99_latency"`
	MinLatency      time.Duration `json:"min_latency"`
	MaxLatency      time.Duration `json:"max_latency"`
	ActualRPS       float64       `json:"actual_rps"`
	ErrorRate       float64       `json:"error_rate"`
	TestDuration    time.Duration `json:"test_duration"`
	MemoryUsage     MemoryStats   `json:"memory_usage"`
	ConcurrentConns int           `json:"concurrent_connections"`
	ResponseCodes   map[int]int64 `json:"response_codes"`
	PerformanceTier string        `json:"performance_tier"`
}

// MemoryStats holds memory usage statistics
type MemoryStats struct {
	HeapAlloc     uint64 `json:"heap_alloc_mb"`
	HeapSys       uint64 `json:"heap_sys_mb"`
	HeapInuse     uint64 `json:"heap_inuse_mb"`
	StackInuse    uint64 `json:"stack_inuse_mb"`
	NumGC         uint32 `json:"num_gc"`
	PauseTotalNs  uint64 `json:"pause_total_ns"`
	NumGoroutines int    `json:"num_goroutines"`
}

// LatencyTracker tracks request latencies
type LatencyTracker struct {
	latencies []time.Duration
	mutex     sync.RWMutex
}

func (lt *LatencyTracker) Add(latency time.Duration) {
	lt.mutex.Lock()
	defer lt.mutex.Unlock()
	lt.latencies = append(lt.latencies, latency)
}

func (lt *LatencyTracker) GetStats() (min, max, avg, p95, p99 time.Duration) {
	lt.mutex.RLock()
	defer lt.mutex.RUnlock()

	if len(lt.latencies) == 0 {
		return 0, 0, 0, 0, 0
	}

	// Sort latencies for percentile calculation
	sorted := make([]time.Duration, len(lt.latencies))
	copy(sorted, lt.latencies)

	// Simple bubble sort for simplicity
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	min = sorted[0]
	max = sorted[len(sorted)-1]

	// Calculate average
	var total time.Duration
	for _, lat := range sorted {
		total += lat
	}
	avg = total / time.Duration(len(sorted))

	// Calculate percentiles
	p95Index := int(float64(len(sorted)) * 0.95)
	p99Index := int(float64(len(sorted)) * 0.99)

	if p95Index >= len(sorted) {
		p95Index = len(sorted) - 1
	}
	if p99Index >= len(sorted) {
		p99Index = len(sorted) - 1
	}

	p95 = sorted[p95Index]
	p99 = sorted[p99Index]

	return min, max, avg, p95, p99
}

// RequestJob represents a single request job for the worker pool
type RequestJob struct {
	Endpoint  string
	AuthToken string
}

// PerformanceTest represents a single performance test
type PerformanceTest struct {
	config          TestConfig
	client          *http.Client
	latencyTracker  *LatencyTracker
	totalRequests   int64
	successfulReqs  int64
	failedReqs      int64
	responseCodes   map[int]int64
	responseCodesMu sync.RWMutex
	startTime       time.Time
	endTime         time.Time
	activeConns     int64
	jobQueue        chan RequestJob
	workerWg        sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
}

func NewPerformanceTest(config TestConfig) *PerformanceTest {
	// Set default worker pool size if not specified
	if config.WorkerPoolSize == 0 {
		config.WorkerPoolSize = 100 // Conservative default
	}

	// Create HTTP client with connection pooling
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        config.MaxConcurrency,
			MaxIdleConnsPerHost: config.MaxConcurrency,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  true, // Reduce CPU overhead
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &PerformanceTest{
		config:         config,
		client:         client,
		latencyTracker: &LatencyTracker{},
		responseCodes:  make(map[int]int64),
		jobQueue:       make(chan RequestJob, config.WorkerPoolSize*2), // Buffer for job queue
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (pt *PerformanceTest) generateAuthToken() string {
	if !pt.config.UseAuth {
		return ""
	}

	switch pt.config.AuthType {
	case "jwt", "both":
		// Generate a test JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":  pt.config.Username,
			"exp":  time.Now().Add(time.Hour).Unix(),
			"iat":  time.Now().Unix(),
			"role": pt.config.UserRole,
		})

		tokenString, _ := token.SignedString([]byte(pt.config.JWTSecret))
		return tokenString
	default:
		return ""
	}
}

// startWorkers starts the worker pool
func (pt *PerformanceTest) startWorkers() {
	for i := 0; i < pt.config.WorkerPoolSize; i++ {
		pt.workerWg.Add(1)
		go pt.worker()
	}
}

// worker processes jobs from the job queue
func (pt *PerformanceTest) worker() {
	defer pt.workerWg.Done()

	for {
		select {
		case <-pt.ctx.Done():
			return
		case job, ok := <-pt.jobQueue:
			if !ok {
				return
			}
			pt.makeRequest(job.Endpoint, job.AuthToken)
		}
	}
}

func (pt *PerformanceTest) makeRequest(endpoint string, authToken string) {
	atomic.AddInt64(&pt.activeConns, 1)
	defer atomic.AddInt64(&pt.activeConns, -1)

	start := time.Now()
	defer func() {
		latency := time.Since(start)
		pt.latencyTracker.Add(latency)
	}()

	req, err := http.NewRequest("GET", pt.config.BaseURL+endpoint, nil)
	if err != nil {
		atomic.AddInt64(&pt.failedReqs, 1)
		atomic.AddInt64(&pt.totalRequests, 1)
		return
	}

	// Set authentication headers based on configuration
	if pt.config.UseAuth {
		switch pt.config.AuthType {
		case "jwt":
			if authToken != "" {
				req.Header.Set("Authorization", "Bearer "+authToken)
			}
		case "api-key":
			if pt.config.APIKey != "" {
				req.Header.Set(pt.config.APIKeyHeader, pt.config.APIKey)
			}
		case "both":
			// Try JWT first, then API key
			if authToken != "" {
				req.Header.Set("Authorization", "Bearer "+authToken)
			} else if pt.config.APIKey != "" {
				req.Header.Set(pt.config.APIKeyHeader, pt.config.APIKey)
			}
		}
	} else if authToken != "" {
		// Fallback for backward compatibility
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	req.Header.Set("User-Agent", "PerformanceTest/1.0")

	resp, err := pt.client.Do(req)
	if err != nil {
		atomic.AddInt64(&pt.failedReqs, 1)
		atomic.AddInt64(&pt.totalRequests, 1)
		return
	}
	defer resp.Body.Close()

	// Track response codes
	pt.responseCodesMu.Lock()
	pt.responseCodes[resp.StatusCode]++
	pt.responseCodesMu.Unlock()

	atomic.AddInt64(&pt.totalRequests, 1)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		atomic.AddInt64(&pt.successfulReqs, 1)
	} else {
		atomic.AddInt64(&pt.failedReqs, 1)
	}
}

func (pt *PerformanceTest) runRampUpTest() TestResult {
	fmt.Println("Starting ramp-up performance test...")

	pt.startTime = time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), pt.config.Duration)
	defer cancel()

	// Start worker pool
	pt.startWorkers()

	// Warmup phase
	fmt.Printf("Warming up for %v...\n", pt.config.WarmupDuration)
	pt.warmup()

	// Reset counters after warmup
	atomic.StoreInt64(&pt.totalRequests, 0)
	atomic.StoreInt64(&pt.successfulReqs, 0)
	atomic.StoreInt64(&pt.failedReqs, 0)
	pt.latencyTracker = &LatencyTracker{}
	pt.responseCodes = make(map[int]int64)

	// Main test phase
	fmt.Println("Starting main performance test...")
	pt.startTime = time.Now()

	// Calculate RPS ramp-up
	rpsIncrement := float64(pt.config.MaxRPS-pt.config.InitialRPS) / pt.config.RampUpDuration.Seconds()

	authToken := pt.generateAuthToken()

	// Use a ticker for controlled request rate
	ticker := time.NewTicker(100 * time.Millisecond) // Check every 100ms
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				elapsed := time.Since(pt.startTime)
				if elapsed > pt.config.Duration {
					return
				}

				// Calculate current target RPS
				var currentRPS int
				if elapsed <= pt.config.RampUpDuration {
					currentRPS = pt.config.InitialRPS + int(float64(elapsed.Seconds())*rpsIncrement)
				} else {
					currentRPS = pt.config.MaxRPS
				}

				// Calculate requests per tick (100ms)
				requestsPerTick := float64(currentRPS) / 10.0 // 10 ticks per second

				// Launch requests using worker pool
				for i := 0; i < int(requestsPerTick); i++ {
					select {
					case <-ctx.Done():
						return
					case pt.jobQueue <- RequestJob{
						Endpoint:  pt.config.TestEndpoints[atomic.LoadInt64(&pt.totalRequests)%int64(len(pt.config.TestEndpoints))],
						AuthToken: authToken,
					}:
						// Job queued successfully
					default:
						// Queue is full, skip this request to prevent blocking
						atomic.AddInt64(&pt.failedReqs, 1)
						atomic.AddInt64(&pt.totalRequests, 1)
					}
				}
			}
		}
	}()

	// Wait for test to complete
	<-ctx.Done()
	pt.endTime = time.Now()

	// Close job queue and wait for workers to finish
	close(pt.jobQueue)
	pt.workerWg.Wait()

	return pt.generateResult()
}

func (pt *PerformanceTest) warmup() {
	authToken := pt.generateAuthToken()

	// Send warmup requests using worker pool
	for i := 0; i < 50; i++ { // Reduced from 100 to 50
		select {
		case pt.jobQueue <- RequestJob{
			Endpoint:  pt.config.TestEndpoints[i%len(pt.config.TestEndpoints)],
			AuthToken: authToken,
		}:
		default:
			// Skip if queue is full
		}
	}

	time.Sleep(pt.config.WarmupDuration)
}

func (pt *PerformanceTest) generateResult() TestResult {
	pt.endTime = time.Now()
	testDuration := pt.endTime.Sub(pt.startTime)

	min, max, avg, p95, p99 := pt.latencyTracker.GetStats()

	totalReqs := atomic.LoadInt64(&pt.totalRequests)
	successfulReqs := atomic.LoadInt64(&pt.successfulReqs)
	failedReqs := atomic.LoadInt64(&pt.failedReqs)

	var errorRate float64
	if totalReqs > 0 {
		errorRate = float64(failedReqs) / float64(totalReqs) * 100
	}

	actualRPS := float64(totalReqs) / testDuration.Seconds()

	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memoryStats := MemoryStats{
		HeapAlloc:     m.HeapAlloc / 1024 / 1024, // Convert to MB
		HeapSys:       m.HeapSys / 1024 / 1024,
		HeapInuse:     m.HeapInuse / 1024 / 1024,
		StackInuse:    m.StackInuse / 1024 / 1024,
		NumGC:         m.NumGC,
		PauseTotalNs:  m.PauseTotalNs,
		NumGoroutines: runtime.NumGoroutine(),
	}

	result := TestResult{
		TotalRequests:   totalReqs,
		SuccessfulReqs:  successfulReqs,
		FailedReqs:      failedReqs,
		AvgLatency:      avg,
		P95Latency:      p95,
		P99Latency:      p99,
		MinLatency:      min,
		MaxLatency:      max,
		ActualRPS:       actualRPS,
		ErrorRate:       errorRate,
		TestDuration:    testDuration,
		MemoryUsage:     memoryStats,
		ConcurrentConns: int(atomic.LoadInt64(&pt.activeConns)),
		ResponseCodes:   pt.responseCodes,
	}

	result.PerformanceTier = pt.determinePerformanceTier(result)
	return result
}

func (pt *PerformanceTest) determinePerformanceTier(result TestResult) string {
	if result.ErrorRate > 5.0 {
		return "Poor"
	}
	if result.AvgLatency > 500*time.Millisecond {
		return "Fair"
	}
	if result.AvgLatency > 200*time.Millisecond {
		return "Good"
	}
	if result.AvgLatency > 100*time.Millisecond {
		return "Very Good"
	}
	return "Excellent"
}

func (pt *PerformanceTest) printResults(result TestResult) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("PERFORMANCE TEST RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("Test Duration: %v\n", result.TestDuration)
	fmt.Printf("Total Requests: %d\n", result.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", result.SuccessfulReqs)
	fmt.Printf("Failed Requests: %d\n", result.FailedReqs)
	fmt.Printf("Error Rate: %.2f%%\n", result.ErrorRate)
	fmt.Printf("Actual RPS: %.2f\n", result.ActualRPS)
	fmt.Printf("Concurrent Connections: %d\n", result.ConcurrentConns)

	fmt.Println("\nLatency Statistics:")
	fmt.Printf("  Average: %v\n", result.AvgLatency)
	fmt.Printf("  95th Percentile: %v\n", result.P95Latency)
	fmt.Printf("  99th Percentile: %v\n", result.P99Latency)
	fmt.Printf("  Minimum: %v\n", result.MinLatency)
	fmt.Printf("  Maximum: %v\n", result.MaxLatency)

	fmt.Println("\nMemory Usage:")
	fmt.Printf("  Heap Allocated: %d MB\n", result.MemoryUsage.HeapAlloc)
	fmt.Printf("  Heap System: %d MB\n", result.MemoryUsage.HeapSys)
	fmt.Printf("  Heap In Use: %d MB\n", result.MemoryUsage.HeapInuse)
	fmt.Printf("  Stack In Use: %d MB\n", result.MemoryUsage.StackInuse)
	fmt.Printf("  Number of GCs: %d\n", result.MemoryUsage.NumGC)
	fmt.Printf("  Total GC Pause: %v\n", time.Duration(result.MemoryUsage.PauseTotalNs))
	fmt.Printf("  Active Goroutines: %d\n", result.MemoryUsage.NumGoroutines)

	fmt.Println("\nResponse Codes:")
	for code, count := range result.ResponseCodes {
		fmt.Printf("  %d: %d\n", code, count)
	}

	fmt.Printf("\nPerformance Tier: %s\n", result.PerformanceTier)
	fmt.Println(strings.Repeat("=", 60))
}
