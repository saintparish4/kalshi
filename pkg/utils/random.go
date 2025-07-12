package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

const (
	// Character sets for random generation
	AlphaLower   = "abcdefghijklmnopqrstuvwxyz"
	AlphaUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Alpha        = AlphaLower + AlphaUpper
	Numeric      = "0123456789"
	Alphanumeric = Alpha + Numeric
	Symbols      = "!@#$%^&*()-_=+[]{}|;:,.<>?"
	AllChars     = Alphanumeric + Symbols
)

// GenerateRandomString generates a random string with specified length and charset
func GenerateRandomString(length int, charset ...string) string {
	charSet := Alphanumeric
	if len(charset) > 0 {
		charSet = charset[0]
	}

	result := make([]byte, length)
	charSetLength := big.NewInt(int64(len(charSet)))

	for i := 0; i < length; i++ {
		randomIndex, _ := rand.Int(rand.Reader, charSetLength)
		result[i] = charSet[randomIndex.Int64()]
	}

	return string(result)
}

// GenerateRandomPassword generates a secure random password
func GenerateRandomPassword(length int, includeSymbols bool) string {
	if length < 4 {
		length = 8 // Minimum secure length
	}

	charset := Alphanumeric
	if includeSymbols {
		charset = AllChars
	}

	password := GenerateRandomString(length, charset)

	// Ensure password contains at least one of each type
	if includeSymbols {
		// Replace some characters to ensure variety
		password = ensurePasswordComplexity(password)
	}

	return password
}

// ensurePasswordComplexity ensures password has required character types
func ensurePasswordComplexity(password string) string {
	runes := []rune(password)
	length := len(runes)

	// Ensure at least one lowercase
	if !hasCharacterType(password, AlphaLower) {
		runes[0] = rune(AlphaLower[randomInt(len(AlphaLower))])
	}

	// Ensure at least one uppercase
	if !hasCharacterType(password, AlphaUpper) {
		runes[1] = rune(AlphaUpper[randomInt(len(AlphaUpper))])
	}

	// Ensure at least one digit
	if !hasCharacterType(password, Numeric) {
		runes[2] = rune(Numeric[randomInt(len(Numeric))])
	}

	// Ensure at least one symbol (if password is long enough)
	if length > 3 && !hasCharacterType(password, Symbols) {
		runes[3] = rune(Symbols[randomInt(len(Symbols))])
	}

	return string(runes)
}

// hasCharacterType checks if string contains characters from charset
func hasCharacterType(s, charset string) bool {
	for _, char := range s {
		if strings.ContainsRune(charset, char) {
			return true
		}
	}
	return false
}

// randomInt generates a cryptographically secure random int
func randomInt(max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}

// GenerateRandomBytes generates random bytes
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	return bytes, err
}

// GenerateUUID generates a UUID v4
func GenerateUUID() string {
	bytes, _ := GenerateRandomBytes(16)

	// Set version (4) and variant bits
	bytes[6] = (bytes[6] & 0x0f) | 0x40 // Version 4
	bytes[8] = (bytes[8] & 0x3f) | 0x80 // Variant 10

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		bytes[0:4],
		bytes[4:6],
		bytes[6:8],
		bytes[8:10],
		bytes[10:16])
}

// GenerateRandomInt generates random int in range [min, max]
func GenerateRandomInt(min, max int) int {
	if min >= max {
		return min
	}

	diff := max - min + 1
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(diff)))
	return min + int(n.Int64())
}

// GenerateRandomFloat generates random float64 in range [min, max)
func GenerateRandomFloat(min, max float64) float64 {
	if min >= max {
		return min
	}

	// Generate random int and convert to float
	precision := 1000000 // 6 decimal places
	minInt := int64(min * float64(precision))
	maxInt := int64(max * float64(precision))

	randomInt, _ := rand.Int(rand.Reader, big.NewInt(maxInt-minInt))
	return float64(minInt+randomInt.Int64()) / float64(precision)
}

// ShuffleSlice shuffles a slice in place
func ShuffleSlice[T any](slice []T) {
	for i := len(slice) - 1; i > 0; i-- {
		j := GenerateRandomInt(0, i)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// RandomChoice returns a random element from slice
func RandomChoice[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}

	index := GenerateRandomInt(0, len(slice)-1)
	return slice[index]
}

// RandomChoices returns n random elements from slice (with replacement)
func RandomChoices[T any](slice []T, n int) []T {
	if len(slice) == 0 {
		return []T{}
	}

	result := make([]T, n)
	for i := 0; i < n; i++ {
		result[i] = RandomChoice(slice)
	}
	return result
}

// RandomSample returns n random elements from slice (without replacement)
func RandomSample[T any](slice []T, n int) []T {
	if n >= len(slice) {
		// Return shuffled copy of entire slice
		result := make([]T, len(slice))
		copy(result, slice)
		ShuffleSlice(result)
		return result
	}

	// Create copy and shuffle
	temp := make([]T, len(slice))
	copy(temp, slice)
	ShuffleSlice(temp)

	return temp[:n]
}

// GenerateOTP generates a numeric OTP (One-Time Password)
func GenerateOTP(length int) string {
	return GenerateRandomString(length, Numeric)
}

// GenerateSessionID generates a secure session ID
func GenerateSessionID() string {
	return GenerateRandomString(32, Alphanumeric)
}

// GenerateCSRFToken generates a CSRF token
func GenerateCSRFToken() string {
	return GenerateRandomString(32, Alphanumeric)
}
