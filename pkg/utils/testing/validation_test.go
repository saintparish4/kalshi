package testing

import (
	"testing"

	"kalshi/pkg/utils"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"valid email", "test@example.com", true},
		{"valid email with subdomain", "test@sub.example.com", true},
		{"invalid email no @", "testexample.com", false},
		{"invalid email no domain", "test@", false},
		{"invalid email no local", "@example.com", false},
		{"empty email", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsValidEmail(tt.email); got != tt.expected {
				t.Errorf("IsValidEmail() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"valid http url", "http://example.com", true},
		{"valid https url", "https://example.com", true},
		{"valid url with path", "https://example.com/path", true},
		{"invalid url", "not-a-url", false},
		{"empty url", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsValidURL(tt.url); got != tt.expected {
				t.Errorf("IsValidURL() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsValidIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"valid IPv4", "192.168.1.1", true},
		{"valid IPv6", "2001:db8::1", true},
		{"invalid IP", "256.256.256.256", false},
		{"empty IP", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsValidIP(tt.ip); got != tt.expected {
				t.Errorf("IsValidIP() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsValidPort(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected bool
	}{
		{"valid port", "8080", true},
		{"valid port min", "1", true},
		{"valid port max", "65535", true},
		{"invalid port 0", "0", false},
		{"invalid port too high", "65536", false},
		{"invalid port string", "abc", false},
		{"empty port", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsValidPort(tt.port); got != tt.expected {
				t.Errorf("IsValidPort() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected bool
	}{
		{"valid username", "user123", true},
		{"valid username with underscore", "user_name", true},
		{"valid username with hyphen", "user-name", true},
		{"too short", "ab", false},
		{"too long", "a" + string(make([]rune, utils.MaxUsernameLength)), false},
		{"invalid chars", "user@name", false},
		{"empty username", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsValidUsername(tt.username); got != tt.expected {
				t.Errorf("IsValidUsername() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"valid password", "Password123!", true},
		{"too short", "Pass1!", false},
		{"no uppercase", "password123!", false},
		{"no lowercase", "PASSWORD123!", false},
		{"no digit", "Password!", false},
		{"no special", "Password123", false},
		{"empty password", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsValidPassword(tt.password); got != tt.expected {
				t.Errorf("IsValidPassword() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected int
	}{
		{"weak password", "pass", 1},
		{"medium password", "Password123!", 4},
		{"strong password", "VeryStrongPassword123!", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.PasswordStrength(tt.password); got != tt.expected {
				t.Errorf("PasswordStrength() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsValidCreditCard(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{"valid visa", "4532015112830366", true},
		{"valid mastercard", "5425233430109903", true},
		{"invalid number", "4532015112830367", false},
		{"too short", "123456789", false},
		{"non-numeric", "453201511283036a", false},
		{"empty number", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsValidCreditCard(tt.number); got != tt.expected {
				t.Errorf("IsValidCreditCard() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsValidJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected bool
	}{
		{"valid json object", `{"key": "value"}`, true},
		{"valid json array", `[1, 2, 3]`, true},
		{"invalid json", `{"key": "value"`, false},
		{"empty json", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsValidJSON(tt.json); got != tt.expected {
				t.Errorf("IsValidJSON() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name     string
		uuid     string
		expected bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", true},
		{"invalid uuid", "not-a-uuid", false},
		{"empty uuid", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsValidUUID(tt.uuid); got != tt.expected {
				t.Errorf("IsValidUUID() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidationResult(t *testing.T) {
	result := &utils.ValidationResult{Valid: true}

	// Test initial state
	if !result.Valid {
		t.Error("ValidationResult should be valid initially")
	}
	if len(result.Errors) != 0 {
		t.Error("ValidationResult should have no errors initially")
	}

	// Test adding error
	result.AddError("test error")
	if result.Valid {
		t.Error("ValidationResult should be invalid after adding error")
	}
	if len(result.Errors) != 1 {
		t.Error("ValidationResult should have one error")
	}
	if result.Errors[0] != "test error" {
		t.Error("ValidationResult should contain the added error")
	}
}

func TestValidateStruct(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		expected bool
	}{
		{"valid string", "test", true},
		{"empty string", "", false},
		{"valid int", 5, true},
		{"invalid int", 0, false},
		{"valid map", map[string]interface{}{"key": "value"}, true},
		{"empty map", map[string]interface{}{}, false},
		{"valid slice", []interface{}{1, 2, 3}, true},
		{"empty slice", []interface{}{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ValidateStruct(tt.data)
			if result.Valid != tt.expected {
				t.Errorf("ValidateStruct() = %v, want %v", result.Valid, tt.expected)
			}
		})
	}
}
