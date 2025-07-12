package testing

import (
	"strings"
	"testing"

	"kalshi/pkg/utils"
)

func TestGenerateRandomString(t *testing.T) {
	// Test with default charset
	result := utils.GenerateRandomString(10)
	if len(result) != 10 {
		t.Errorf("GenerateRandomString(10) length = %d, want 10", len(result))
	}

	// Test with custom charset
	result = utils.GenerateRandomString(5, utils.AlphaLower)
	if len(result) != 5 {
		t.Errorf("GenerateRandomString(5, AlphaLower) length = %d, want 5", len(result))
	}

	// Check if result only contains lowercase letters
	for _, char := range result {
		if !strings.ContainsRune(utils.AlphaLower, char) {
			t.Errorf("GenerateRandomString() with AlphaLower contains invalid character: %c", char)
		}
	}

	// Test with numeric charset
	result = utils.GenerateRandomString(8, utils.Numeric)
	if len(result) != 8 {
		t.Errorf("GenerateRandomString(8, Numeric) length = %d, want 8", len(result))
	}

	// Check if result only contains numbers
	for _, char := range result {
		if !strings.ContainsRune(utils.Numeric, char) {
			t.Errorf("GenerateRandomString() with Numeric contains invalid character: %c", char)
		}
	}
}

func TestGenerateRandomPassword(t *testing.T) {
	// Test with symbols
	password := utils.GenerateRandomPassword(12, true)
	if len(password) != 12 {
		t.Errorf("GenerateRandomPassword(12, true) length = %d, want 12", len(password))
	}

	// Test without symbols
	password = utils.GenerateRandomPassword(10, false)
	if len(password) != 10 {
		t.Errorf("GenerateRandomPassword(10, false) length = %d, want 10", len(password))
	}

	// Test minimum length enforcement
	password = utils.GenerateRandomPassword(2, true)
	if len(password) != 8 {
		t.Errorf("GenerateRandomPassword(2, true) length = %d, want 8 (minimum)", len(password))
	}

	// Test password complexity with symbols
	password = utils.GenerateRandomPassword(8, true)
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSymbol := false

	for _, char := range password {
		if strings.ContainsRune(utils.AlphaLower, char) {
			hasLower = true
		}
		if strings.ContainsRune(utils.AlphaUpper, char) {
			hasUpper = true
		}
		if strings.ContainsRune(utils.Numeric, char) {
			hasDigit = true
		}
		if strings.ContainsRune(utils.Symbols, char) {
			hasSymbol = true
		}
	}

	if !hasLower || !hasUpper || !hasDigit || !hasSymbol {
		t.Error("GenerateRandomPassword() with symbols should contain all character types")
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	// Test valid length
	bytes, err := utils.GenerateRandomBytes(16)
	if err != nil {
		t.Errorf("GenerateRandomBytes(16) unexpected error: %v", err)
	}
	if len(bytes) != 16 {
		t.Errorf("GenerateRandomBytes(16) length = %d, want 16", len(bytes))
	}

	// Test zero length
	bytes, err = utils.GenerateRandomBytes(0)
	if err != nil {
		t.Errorf("GenerateRandomBytes(0) unexpected error: %v", err)
	}
	if len(bytes) != 0 {
		t.Errorf("GenerateRandomBytes(0) length = %d, want 0", len(bytes))
	}
}

func TestGenerateUUID(t *testing.T) {
	uuid := utils.GenerateUUID()

	// Check UUID format (8-4-4-4-12 characters)
	parts := strings.Split(uuid, "-")
	if len(parts) != 5 {
		t.Errorf("GenerateUUID() format invalid: %s", uuid)
	}

	if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 || len(parts[3]) != 4 || len(parts[4]) != 12 {
		t.Errorf("GenerateUUID() format invalid: %s", uuid)
	}

	// Check if all characters are hexadecimal
	uuidWithoutHyphens := strings.ReplaceAll(uuid, "-", "")
	for _, char := range uuidWithoutHyphens {
		if !strings.ContainsRune("0123456789abcdef", char) {
			t.Errorf("GenerateUUID() contains non-hex character: %c", char)
		}
	}

	// Test that UUIDs are different
	uuid2 := utils.GenerateUUID()
	if uuid == uuid2 {
		t.Error("GenerateUUID() should generate different UUIDs")
	}
}

func TestGenerateRandomInt(t *testing.T) {
	// Test normal range
	result := utils.GenerateRandomInt(1, 10)
	if result < 1 || result > 10 {
		t.Errorf("GenerateRandomInt(1, 10) = %d, want between 1 and 10", result)
	}

	// Test same min and max
	result = utils.GenerateRandomInt(5, 5)
	if result != 5 {
		t.Errorf("GenerateRandomInt(5, 5) = %d, want 5", result)
	}

	// Test invalid range (min > max)
	result = utils.GenerateRandomInt(10, 1)
	if result != 10 {
		t.Errorf("GenerateRandomInt(10, 1) = %d, want 10", result)
	}
}

func TestGenerateRandomFloat(t *testing.T) {
	// Test normal range
	result := utils.GenerateRandomFloat(0.0, 1.0)
	if result < 0.0 || result >= 1.0 {
		t.Errorf("GenerateRandomFloat(0.0, 1.0) = %f, want between 0.0 and 1.0", result)
	}

	// Test same min and max
	result = utils.GenerateRandomFloat(5.5, 5.5)
	if result != 5.5 {
		t.Errorf("GenerateRandomFloat(5.5, 5.5) = %f, want 5.5", result)
	}

	// Test invalid range (min > max)
	result = utils.GenerateRandomFloat(1.0, 0.0)
	if result != 1.0 {
		t.Errorf("GenerateRandomFloat(1.0, 0.0) = %f, want 1.0", result)
	}
}

func TestShuffleSlice(t *testing.T) {
	original := []int{1, 2, 3, 4, 5}
	shuffled := make([]int, len(original))
	copy(shuffled, original)

	utils.ShuffleSlice(shuffled)

	// Check that all elements are still present
	if len(shuffled) != len(original) {
		t.Errorf("ShuffleSlice() changed slice length from %d to %d", len(original), len(shuffled))
	}

	// Check that elements are the same (order may be different)
	originalMap := make(map[int]int)
	shuffledMap := make(map[int]int)

	for _, v := range original {
		originalMap[v]++
	}
	for _, v := range shuffled {
		shuffledMap[v]++
	}

	for k, v := range originalMap {
		if shuffledMap[k] != v {
			t.Errorf("ShuffleSlice() changed element count for %d", k)
		}
	}
}

func TestRandomChoice(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}

	// Test multiple choices to ensure randomness
	choices := make(map[string]int)
	for i := 0; i < 100; i++ {
		choice := utils.RandomChoice(slice)
		choices[choice]++
	}

	// Check that all elements were chosen at least once
	for _, element := range slice {
		if choices[element] == 0 {
			t.Errorf("RandomChoice() never chose element: %s", element)
		}
	}

	// Test empty slice
	emptySlice := []string{}
	result := utils.RandomChoice(emptySlice)
	if result != "" {
		t.Errorf("RandomChoice() with empty slice = %s, want empty string", result)
	}
}

func TestRandomChoices(t *testing.T) {
	slice := []string{"a", "b", "c"}
	n := 5

	result := utils.RandomChoices(slice, n)
	if len(result) != n {
		t.Errorf("RandomChoices() length = %d, want %d", len(result), n)
	}

	// Check that all elements are from the original slice
	for _, choice := range result {
		found := false
		for _, element := range slice {
			if choice == element {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("RandomChoices() returned element not in original slice: %s", choice)
		}
	}

	// Test with empty slice
	emptySlice := []string{}
	result = utils.RandomChoices(emptySlice, 3)
	if len(result) != 0 {
		t.Errorf("RandomChoices() with empty slice = %v, want empty slice", result)
	}
}

func TestRandomSample(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	n := 3

	result := utils.RandomSample(slice, n)
	if len(result) != n {
		t.Errorf("RandomSample() length = %d, want %d", len(result), n)
	}

	// Check that all elements are from the original slice
	for _, choice := range result {
		found := false
		for _, element := range slice {
			if choice == element {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("RandomSample() returned element not in original slice: %d", choice)
		}
	}

	// Check that no element is repeated
	seen := make(map[int]bool)
	for _, choice := range result {
		if seen[choice] {
			t.Errorf("RandomSample() returned duplicate element: %d", choice)
		}
		seen[choice] = true
	}

	// Test with n >= len(slice)
	result = utils.RandomSample(slice, 10)
	if len(result) != len(slice) {
		t.Errorf("RandomSample() with n >= len(slice) = %d, want %d", len(result), len(slice))
	}
}

func TestGenerateOTP(t *testing.T) {
	otp := utils.GenerateOTP(6)
	if len(otp) != 6 {
		t.Errorf("GenerateOTP(6) length = %d, want 6", len(otp))
	}

	// Check that OTP only contains digits
	for _, char := range otp {
		if char < '0' || char > '9' {
			t.Errorf("GenerateOTP() contains non-digit character: %c", char)
		}
	}

	// Test different lengths
	otp = utils.GenerateOTP(4)
	if len(otp) != 4 {
		t.Errorf("GenerateOTP(4) length = %d, want 4", len(otp))
	}
}

func TestGenerateSessionID(t *testing.T) {
	sessionID := utils.GenerateSessionID()
	if len(sessionID) == 0 {
		t.Error("GenerateSessionID() returned empty string")
	}

	// Session IDs should be different
	sessionID2 := utils.GenerateSessionID()
	if sessionID == sessionID2 {
		t.Error("GenerateSessionID() should generate different session IDs")
	}
}

func TestGenerateCSRFToken(t *testing.T) {
	token := utils.GenerateCSRFToken()
	if len(token) == 0 {
		t.Error("GenerateCSRFToken() returned empty string")
	}

	// CSRF tokens should be different
	token2 := utils.GenerateCSRFToken()
	if token == token2 {
		t.Error("GenerateCSRFToken() should generate different tokens")
	}
}
