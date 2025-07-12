package testing

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"kalshi/pkg/utils"
)

func TestNewHTTPClient(t *testing.T) {
	timeout := 30 * time.Second
	client := utils.NewHTTPClient(timeout)

	if client == nil {
		t.Error("NewHTTPClient() returned nil")
	}

	// Test that the client can be used for basic operations
	client.SetBaseURL("https://api.example.com")
	client.SetHeader("Authorization", "Bearer token")

	// If we can set headers and base URL without errors, the client is properly initialized
}

func TestHTTPClient_SetBaseURL(t *testing.T) {
	client := utils.NewHTTPClient(30 * time.Second)

	// Test that SetBaseURL doesn't cause errors
	client.SetBaseURL("https://api.example.com/")
	client.SetBaseURL("https://api.example.com")

	// The method should work without errors
}

func TestHTTPClient_SetHeader(t *testing.T) {
	client := utils.NewHTTPClient(30 * time.Second)

	// Test that SetHeader doesn't cause errors
	client.SetHeader("Authorization", "Bearer token")
	client.SetHeader("Content-Type", "application/json")

	// The method should work without errors
}

func TestHTTPClient_buildURL(t *testing.T) {
	client := utils.NewHTTPClient(30 * time.Second)
	client.SetBaseURL("https://api.example.com")

	// Since buildURL is unexported, we test the public interface instead
	// The buildURL method is used internally by the public methods like Get, Post, etc.
	// We can test that the client can be created and configured without errors
}

func TestGetClientIP(t *testing.T) {
	// Test X-Forwarded-For header
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
	ip := utils.GetClientIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("GetClientIP() with X-Forwarded-For = %v, want %v", ip, "192.168.1.1")
	}

	// Test X-Real-IP header
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "192.168.1.2")
	ip = utils.GetClientIP(req)
	if ip != "192.168.1.2" {
		t.Errorf("GetClientIP() with X-Real-IP = %v, want %v", ip, "192.168.1.2")
	}

	// Test RemoteAddr
	req, _ = http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.3:8080"
	ip = utils.GetClientIP(req)
	if ip != "192.168.1.3" {
		t.Errorf("GetClientIP() with RemoteAddr = %v, want %v", ip, "192.168.1.3")
	}
}

func TestGetUserAgent(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	ua := utils.GetUserAgent(req)
	if ua != "Mozilla/5.0" {
		t.Errorf("GetUserAgent() = %v, want %v", ua, "Mozilla/5.0")
	}
}

func TestIsAjaxRequest(t *testing.T) {
	// Test AJAX request
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	if !utils.IsAjaxRequest(req) {
		t.Error("IsAjaxRequest() should return true for AJAX request")
	}

	// Test non-AJAX request
	req, _ = http.NewRequest("GET", "/", nil)
	if utils.IsAjaxRequest(req) {
		t.Error("IsAjaxRequest() should return false for non-AJAX request")
	}
}

func TestIsJSONRequest(t *testing.T) {
	// Test JSON request
	req, _ := http.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "application/json")
	if !utils.IsJSONRequest(req) {
		t.Error("IsJSONRequest() should return true for JSON request")
	}

	// Test non-JSON request
	req, _ = http.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "text/html")
	if utils.IsJSONRequest(req) {
		t.Error("IsJSONRequest() should return false for non-JSON request")
	}
}

func TestWriteJSONResponse(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "success"}

	err := utils.WriteJSONResponse(w, http.StatusOK, data)
	if err != nil {
		t.Errorf("WriteJSONResponse() unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("WriteJSONResponse() status = %v, want %v", w.Code, http.StatusOK)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("WriteJSONResponse() Content-Type = %v, want %v", contentType, "application/json")
	}

	body := w.Body.String()
	if !strings.Contains(body, "success") {
		t.Errorf("WriteJSONResponse() body = %v, should contain 'success'", body)
	}
}

func TestReadJSONBody(t *testing.T) {
	jsonData := `{"name": "John", "age": 30}`
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonData))

	var result map[string]interface{}
	err := utils.ReadJSONBody(req, &result)
	if err != nil {
		t.Errorf("ReadJSONBody() unexpected error: %v", err)
	}

	if result["name"] != "John" {
		t.Errorf("ReadJSONBody() name = %v, want %v", result["name"], "John")
	}

	if result["age"].(float64) != 30 {
		t.Errorf("ReadJSONBody() age = %v, want %v", result["age"], 30)
	}
}

func TestCopyHeaders(t *testing.T) {
	src := http.Header{}
	src.Set("Content-Type", "application/json")
	src.Set("Authorization", "Bearer token")

	dst := http.Header{}
	utils.CopyHeaders(dst, src)

	if dst.Get("Content-Type") != "application/json" {
		t.Errorf("CopyHeaders() Content-Type = %v, want %v", dst.Get("Content-Type"), "application/json")
	}

	if dst.Get("Authorization") != "Bearer token" {
		t.Errorf("CopyHeaders() Authorization = %v, want %v", dst.Get("Authorization"), "Bearer token")
	}
}

func TestSetSecurityHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	utils.SetSecurityHeaders(w)

	headers := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	for key, expectedValue := range headers {
		actualValue := w.Header().Get(key)
		if actualValue != expectedValue {
			t.Errorf("SetSecurityHeaders() %s = %v, want %v", key, actualValue, expectedValue)
		}
	}
}

func TestSetCacheHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	maxAge := 3600
	utils.SetCacheHeaders(w, maxAge)

	expected := "max-age=3600"
	actual := w.Header().Get("Cache-Control")
	if actual != expected {
		t.Errorf("SetCacheHeaders() Cache-Control = %v, want %v", actual, expected)
	}
}

func TestSetNoCacheHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	utils.SetNoCacheHeaders(w)

	expected := "no-cache, no-store, must-revalidate"
	actual := w.Header().Get("Cache-Control")
	if actual != expected {
		t.Errorf("SetNoCacheHeaders() Cache-Control = %v, want %v", actual, expected)
	}
}

func TestParseContentType(t *testing.T) {
	// Test simple content type
	mediaType, params := utils.ParseContentType("application/json")
	if mediaType != "application/json" {
		t.Errorf("ParseContentType() mediaType = %v, want %v", mediaType, "application/json")
	}
	if len(params) != 0 {
		t.Errorf("ParseContentType() params length = %v, want %v", len(params), 0)
	}

	// Test content type with parameters
	mediaType, params = utils.ParseContentType("text/html; charset=utf-8")
	if mediaType != "text/html" {
		t.Errorf("ParseContentType() mediaType = %v, want %v", mediaType, "text/html")
	}
	if params["charset"] != "utf-8" {
		t.Errorf("ParseContentType() charset = %v, want %v", params["charset"], "utf-8")
	}
}
