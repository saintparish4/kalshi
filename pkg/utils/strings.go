package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// IsEmpty checks if string is empty or contains only whitespace
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// Truncate truncates string to specified length with optional suffix
func Truncate(s string, length int, suffix string) string {
	if len(s) <= length {
		return s
	}
	
	if len(suffix) >= length {
		return suffix[:length]
	}
	
	return s[:length-len(suffix)] + suffix
}

// TruncateWords truncates string to specified number of words
func TruncateWords(s string, wordCount int, suffix string) string {
	words := strings.Fields(s)
	if len(words) <= wordCount {
		return s
	}
	
	truncated := strings.Join(words[:wordCount], " ")
	return truncated + suffix
}

// CamelToSnake converts camelCase to snake_case
func CamelToSnake(s string) string {
	var result []rune
	
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	
	return string(result)
}

// SnakeToCamel converts snake_case to camelCase
func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) == 0 {
		return s
	}
	
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	
	return result
}

// TitleCase converts string to Title Case
func TitleCase(s string) string {
	return strings.Title(strings.ToLower(s))
}

// CleanString removes special characters and extra spaces
func CleanString(s string) string {
	// Remove special characters except alphanumeric, spaces, and common punctuation
	reg := regexp.MustCompile(`[^\w\s\-_.,!?]+`)
	cleaned := reg.ReplaceAllString(s, "")
	
	// Replace multiple spaces with single space
	spaceReg := regexp.MustCompile(`\s+`)
	cleaned = spaceReg.ReplaceAllString(cleaned, " ")
	
	return strings.TrimSpace(cleaned)
}

// ExtractDomain extracts domain from email or URL
func ExtractDomain(input string) string {
	// Handle email
	if strings.Contains(input, "@") {
		parts := strings.Split(input, "@")
		if len(parts) == 2 {
			return parts[1]
		}
	}
	
	// Handle URL
	if strings.Contains(input, "://") {
		parts := strings.Split(input, "://")
		if len(parts) > 1 {
			domain := strings.Split(parts[1], "/")[0]
			return strings.Split(domain, ":")[0] // Remove port if present
		}
	}
	
	return input
}

// Slugify converts string to URL-friendly slug
func Slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)
	
	// Replace spaces and underscores with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	
	// Remove special characters
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	s = reg.ReplaceAllString(s, "")
	
	// Remove multiple consecutive hyphens
	hyphenReg := regexp.MustCompile(`-+`)
	s = hyphenReg.ReplaceAllString(s, "-")
	
	// Trim hyphens from start and end
	return strings.Trim(s, "-")
}

// Mask masks sensitive data in strings
func Mask(s string, visibleChars int, maskChar rune) string {
	if len(s) <= visibleChars*2 {
		return strings.Repeat(string(maskChar), len(s))
	}
	
	start := s[:visibleChars]
	end := s[len(s)-visibleChars:]
	middle := strings.Repeat(string(maskChar), len(s)-visibleChars*2)
	
	return start + middle + end
}

// MaskEmail masks email address
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return Mask(email, 2, '*')
	}
	
	username := parts[0]
	domain := parts[1]
	
	if len(username) <= 2 {
		username = strings.Repeat("*", len(username))
	} else {
		username = username[:1] + strings.Repeat("*", len(username)-2) + username[len(username)-1:]
	}
	
	return username + "@" + domain
}

// Contains checks if slice contains string (case-insensitive option)
func Contains(slice []string, item string, caseSensitive bool) bool {
	for _, s := range slice {
		if caseSensitive {
			if s == item {
				return true
			}
		} else {
			if strings.EqualFold(s, item) {
				return true
			}
		}
	}
	return false
}

// RemoveEmptyStrings removes empty strings from slice
func RemoveEmptyStrings(slice []string) []string {
	var result []string
	for _, s := range slice {
		if !IsEmpty(s) {
			result = append(result, s)
		}
	}
	return result
}

// UniqueStrings returns unique strings from slice
func UniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	
	return result
}