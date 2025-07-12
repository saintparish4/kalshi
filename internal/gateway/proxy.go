package gateway

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"kalshi/internal/cache"
	"kalshi/internal/circuit"
	"kalshi/internal/config"
	"kalshi/pkg/logger"
)

type CachedResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

type Proxy struct {
	backendManager *BackendManager
	cacheManager   cache.Cache
	circuitManager *circuit.Manager
	logger         *logger.Logger
	// Phase 1: Optimized HTTP client with connection pooling
	optimizedClient *http.Client
	clientOnce      sync.Once
	config          *config.Config
}

func NewProxy(bm *BackendManager, cm cache.Cache, circuitManager *circuit.Manager, logger *logger.Logger, cfg *config.Config) *Proxy {
	return &Proxy{
		backendManager: bm,
		cacheManager:   cm,
		circuitManager: circuitManager,
		logger:         logger,
		config:         cfg,
	}
}

// getOptimizedClient creates and returns an optimized HTTP client with connection pooling
func (p *Proxy) getOptimizedClient() *http.Client {
	p.clientOnce.Do(func() {
		// Use configuration if available, otherwise use defaults
		var perfConfig config.PerformanceConfig
		if p.config != nil {
			perfConfig = p.config.Performance
		} else {
			// Default high-performance settings
			perfConfig = config.PerformanceConfig{
				MaxIdleConns:          1000,
				MaxIdleConnsPerHost:   1000,
				IdleConnTimeout:       90 * time.Second,
				DisableCompression:    true,
				DisableKeepAlives:     false,
				ForceHTTP2:            true,
				ConnectionTimeout:     5 * time.Second,
				KeepAlive:             30 * time.Second,
				MaxConnsPerHost:       2000,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			}
		}

		// Create optimized transport with connection pooling
		transport := &http.Transport{
			MaxIdleConns:        perfConfig.MaxIdleConns,
			MaxIdleConnsPerHost: perfConfig.MaxIdleConnsPerHost,
			IdleConnTimeout:     perfConfig.IdleConnTimeout,
			DisableCompression:  perfConfig.DisableCompression,
			DisableKeepAlives:   perfConfig.DisableKeepAlives,
			ForceAttemptHTTP2:   perfConfig.ForceHTTP2,

			// Windows-specific optimizations
			DialContext: (&net.Dialer{
				Timeout:   perfConfig.ConnectionTimeout,
				KeepAlive: perfConfig.KeepAlive,
			}).DialContext,

			// High-concurrency optimizations
			MaxConnsPerHost:       perfConfig.MaxConnsPerHost,
			TLSHandshakeTimeout:   perfConfig.TLSHandshakeTimeout,
			ExpectContinueTimeout: perfConfig.ExpectContinueTimeout,
		}

		p.optimizedClient = &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		}
	})

	return p.optimizedClient
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request, backendName string, cacheTTL time.Duration) {
	// Check cache first for GET requests
	if r.Method == "GET" && cacheTTL > 0 {
		if cached, err := p.getCachedResponse(r); err == nil {
			p.writeCachedResponse(w, cached)
			return
		}
	}

	// Get Backend
	backend, err := p.backendManager.GetBackend(backendName)
	if err != nil {
		http.Error(w, "Backend not found", http.StatusBadGateway)
		return
	}

	// Get circuit breaker
	breaker := p.circuitManager.GetBreaker(
		backendName,
		5,              // failure threshold
		30*time.Second, // recovery timeout
		3,              // max requests in half open state
	)

	// Use circuit breaker
	var resp *http.Response
	err = breaker.Call(func() error {
		var callErr error
		resp, callErr = p.forwardRequest(r, backend)
		if callErr != nil {
			return callErr
		}

		// Consider only 5xx HTTP error status codes as failures for circuit breaker
		// 4xx responses are client errors and should be forwarded
		if resp.StatusCode >= 500 {
			return fmt.Errorf("HTTP error: %d", resp.StatusCode)
		}

		return nil
	})

	if err != nil {
		if err == circuit.ErrCircuitBreakerOpen {
			http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
		} else {
			http.Error(w, "Backend error", http.StatusBadGateway)
		}
		return
	}

	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logger.Error("Failed to read response body", "error", err)
		return
	}

	// Cache successful GET responses
	if r.Method == "GET" && resp.StatusCode == 200 && cacheTTL > 0 {
		p.cacheResponse(r, resp, body, cacheTTL)
	}

	// Write response
	w.Write(body)
}

func (p *Proxy) generateCacheKey(r *http.Request) string {
	return fmt.Sprintf("%s:%s", r.Method, r.URL.String())
}

// GenerateCacheKey is a public wrapper for generateCacheKey for testing
func (p *Proxy) GenerateCacheKey(r *http.Request) string {
	return p.generateCacheKey(r)
}

func (p *Proxy) getCachedResponse(r *http.Request) (*CachedResponse, error) {
	cacheKey := p.generateCacheKey(r)

	data, err := p.cacheManager.Get(r.Context(), cacheKey)
	if err != nil {
		return nil, err
	}

	// In a real implementation, you would properly serialize/deserialize the response
	// For simplicity, well just cache the body
	return &CachedResponse{
		StatusCode: 200,
		Headers:    make(http.Header),
		Body:       data,
	}, nil
}

// GetCachedResponse is a public wrapper for getCachedResponse for testing
func (p *Proxy) GetCachedResponse(r *http.Request) (*CachedResponse, error) {
	return p.getCachedResponse(r)
}

func (p *Proxy) writeCachedResponse(w http.ResponseWriter, cached *CachedResponse) {
	for key, values := range cached.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(cached.StatusCode)
	w.Write(cached.Body)
}

// WriteCachedResponse is a public wrapper for writeCachedResponse for testing
func (p *Proxy) WriteCachedResponse(w http.ResponseWriter, cached *CachedResponse) {
	p.writeCachedResponse(w, cached)
}

func (p *Proxy) cacheResponse(r *http.Request, _ *http.Response, body []byte, ttl time.Duration) {
	cacheKey := p.generateCacheKey(r)
	p.cacheManager.Set(r.Context(), cacheKey, body, ttl)
}

// CacheResponse is a public wrapper for cacheResponse for testing
func (p *Proxy) CacheResponse(r *http.Request, resp *http.Response, body []byte, ttl time.Duration) {
	p.cacheResponse(r, resp, body, ttl)
}

func (p *Proxy) forwardRequest(r *http.Request, backend *Backend) (*http.Response, error) {
	// Create target URL
	targetURL := &url.URL{
		Scheme:   backend.URL.Scheme,
		Host:     backend.URL.Host,
		Path:     backend.URL.Path + r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	// Create new request
	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL.String(), r.Body)
	if err != nil {
		return nil, err
	}

	// Copy Headers
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Add forwarding headers
	proxyReq.Header.Add("X-Forwarded-For", r.RemoteAddr)
	proxyReq.Header.Add("X-Forwarded-Proto", r.URL.Scheme)
	proxyReq.Header.Add("X-Forwarded-Host", r.Host)

	// Make Request
	client := p.getOptimizedClient()

	return client.Do(proxyReq)
}

// ForwardRequest is a public wrapper for forwardRequest for testing
func (p *Proxy) ForwardRequest(r *http.Request, backend *Backend) (*http.Response, error) {
	return p.forwardRequest(r, backend)
}

// GetOptimizedClient is a public wrapper for getOptimizedClient for testing
func (p *Proxy) GetOptimizedClient() *http.Client {
	return p.getOptimizedClient()
}
