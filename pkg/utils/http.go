package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPClient is a wrapper around http.Client with additional utilities
type HTTPClient struct {
	client  *http.Client
	baseURL string
	headers map[string]string
}

// NewHTTPClient creates a new HTTP client with default settings
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		headers: make(map[string]string),
	}
}

// SetBaseURL sets the base URL for all requests
func (c *HTTPClient) SetBaseURL(baseURL string) {
	c.baseURL = strings.TrimSuffix(baseURL, "/")
}

// SetHeader sets a default header for all requests
func (c *HTTPClient) SetHeader(key, value string) {
	c.headers[key] = value
}

// Get performs a GET request
func (c *HTTPClient) Get(path string, params map[string]string) (*http.Response, error) {
	url := c.buildURL(path, params)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.setDefaultHeaders(req)
	return c.client.Do(req)
}

// Post performs a POST request with JSON body
func (c *HTTPClient) Post(path string, body interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := c.buildURL(path, nil)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	c.setDefaultHeaders(req)
	return c.client.Do(req)
}

// Put performs a PUT request with JSON body
func (c *HTTPClient) Put(path string, body interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := c.buildURL(path, nil)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	c.setDefaultHeaders(req)
	return c.client.Do(req)
}

// Delete performs a DELETE request
func (c *HTTPClient) Delete(path string) (*http.Response, error) {
	url := c.buildURL(path, nil)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	c.setDefaultHeaders(req)
	return c.client.Do(req)
}

// buildURL constructs the full URL with query parameters
func (c *HTTPClient) buildURL(path string, params map[string]string) string {
	fullURL := c.baseURL + "/" + strings.TrimPrefix(path, "/")

	if len(params) > 0 {
		values := url.Values{}
		for key, value := range params {
			values.Add(key, value)
		}
		fullURL += "?" + values.Encode()
	}

	return fullURL
}

// setDefaultHeaders adds default headers to request
func (c *HTTPClient) setDefaultHeaders(req *http.Request) {
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
}

// GetClientIP extracts client IP from request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}

	return ip
}

// GetUserAgent extracts user agent from request
func GetUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

// IsAjaxRequest checks if request is AJAX
func IsAjaxRequest(r *http.Request) bool {
	return r.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

// IsJSONRequest checks if request content type is JSON
func IsJSONRequest(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	return strings.Contains(contentType, "application/json")
}

// WriteJSONResponse writes JSON response
func WriteJSONResponse(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// ReadJSONBody reads and parses JSON request body
func ReadJSONBody(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.Unmarshal(body, v)
}

// CopyHeaders copies headers from source to destination
func CopyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// SetSecurityHeaders sets common security headers
func SetSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
}

// SetCacheHeaders sets cache control headers
func SetCacheHeaders(w http.ResponseWriter, maxAge int) {
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", maxAge))
	w.Header().Set("Expires", time.Now().Add(time.Duration(maxAge)*time.Second).Format(http.TimeFormat))
}

// SetNoCacheHeaders sets no-cache headers
func SetNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

// ParseContentType parses content type header
func ParseContentType(contentType string) (mediaType string, params map[string]string) {
	parts := strings.Split(contentType, ";")
	mediaType = strings.TrimSpace(parts[0])
	params = make(map[string]string)

	for i := 1; i < len(parts); i++ {
		param := strings.TrimSpace(parts[i])
		if eq := strings.Index(param, "="); eq != -1 {
			key := strings.TrimSpace(param[:eq])
			value := strings.Trim(strings.TrimSpace(param[eq+1:]), "\"")
			params[key] = value
		}
	}

	return mediaType, params
}
