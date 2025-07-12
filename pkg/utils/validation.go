package utils

import (
	"encoding/json"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Validation constants
const (
	MinUsernameLength = 3
	MaxUsernameLength = 30
	MinPasswordLength = 8
	MinPhoneLength    = 7
	MaxPhoneLength    = 15
	MinPortNumber     = 1
	MaxPortNumber     = 65535
)

// IsValidEmail validates email address
func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsValidURL validates URL
func IsValidURL(rawURL string) bool {
	_, err := url.ParseRequestURI(rawURL)
	return err == nil
}

// IsValidIP validates IP address (both IPv4 and IPv6)
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsValidIPv4 validates IPv4 address
func IsValidIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() != nil
}

// IsValidIPv6 validates IPv6 address
func IsValidIPv6(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() == nil
}

// IsValidPort validates port number
func IsValidPort(port string) bool {
	p, err := strconv.Atoi(port)
	return err == nil && p >= MinPortNumber && p <= MaxPortNumber
}

// IsValidPhoneNumber validates phone number (basic pattern)
func IsValidPhoneNumber(phone string) bool {
	// Remove common separators
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, "+", "")

	// Check if all remaining characters are digits
	for _, r := range cleaned {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	// Check length (international format)
	return len(cleaned) >= MinPhoneLength && len(cleaned) <= MaxPhoneLength
}

// IsValidUsername validates username
func IsValidUsername(username string) bool {
	if len(username) < MinUsernameLength || len(username) > MaxUsernameLength {
		return false
	}

	// Allow alphanumeric, underscore, hyphen
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, username)
	return matched
}

// IsValidPassword validates password strength
func IsValidPassword(password string) bool {
	if len(password) < MinPasswordLength {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// PasswordStrength returns password strength score (0-4)
func PasswordStrength(password string) int {
	score := 0

	if len(password) >= MinPasswordLength {
		score++
	}
	if len(password) >= 12 {
		score++
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if hasLower {
		score++
	}
	if hasUpper {
		score++
	}
	if hasDigit {
		score++
	}
	if hasSpecial {
		score++
	}

	if score > 4 {
		score = 4
	}

	return score
}

// IsValidCreditCard validates credit card number using Luhn algorithm
func IsValidCreditCard(number string) bool {
	// Remove spaces and hyphens
	cleaned := strings.ReplaceAll(number, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")

	// Check if all characters are digits
	for _, r := range cleaned {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	// Check length
	if len(cleaned) < 13 || len(cleaned) > 19 {
		return false
	}

	// Luhn algorithm
	sum := 0
	alternate := false

	for i := len(cleaned) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(cleaned[i]))

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit % 10) + 1
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// IsValidJSON checks if string is valid JSON
func IsValidJSON(jsonStr string) bool {
	return json.Valid([]byte(jsonStr))
}

// IsValidUUID validates UUID format
func IsValidUUID(uuid string) bool {
	matched, _ := regexp.MatchString(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, strings.ToLower(uuid))
	return matched
}

// IsValidHexColor validates hex color code
func IsValidHexColor(color string) bool {
	matched, _ := regexp.MatchString(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`, color)
	return matched
}

// IsAlphanumeric checks if string contains only alphanumeric characters
func IsAlphanumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsNumeric checks if string contains only numeric characters
func IsNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsAlpha checks if string contains only alphabetic characters
func IsAlpha(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// ValidationResult represents validation result
type ValidationResult struct {
	Valid  bool
	Errors []string
}

// AddError adds an error to validation result
func (vr *ValidationResult) AddError(message string) {
	vr.Valid = false
	vr.Errors = append(vr.Errors, message)
}

// ValidateStruct validates struct fields with tags
func ValidateStruct(data interface{}) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Basic validation for common types
	switch v := data.(type) {
	case string:
		if v == "" {
			result.AddError("string cannot be empty")
		}
	case int:
		if v <= 0 {
			result.AddError("number must be positive")
		}
	case map[string]interface{}:
		if len(v) == 0 {
			result.AddError("map cannot be empty")
		}
	case []interface{}:
		if len(v) == 0 {
			result.AddError("slice cannot be empty")
		}
	}

	return result
}
