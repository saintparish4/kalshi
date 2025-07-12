package testing

import (
	"testing"

	"kalshi/pkg/utils"
)

func TestEncryptAES(t *testing.T) {
	key := []byte("1234567890123456") // 16 bytes
	data := []byte("test data")

	encrypted, err := utils.EncryptAES(data, key)
	if err != nil {
		t.Errorf("EncryptAES() unexpected error: %v", err)
	}

	if encrypted == "" {
		t.Error("EncryptAES() returned empty string")
	}

	// Test with empty data
	_, err = utils.EncryptAES([]byte{}, key)
	if err != utils.ErrInvalidInput {
		t.Errorf("EncryptAES() with empty data expected ErrInvalidInput, got %v", err)
	}

	// Test with empty key
	_, err = utils.EncryptAES(data, []byte{})
	if err != utils.ErrInvalidInput {
		t.Errorf("EncryptAES() with empty key expected ErrInvalidInput, got %v", err)
	}

	// Test with invalid key size
	_, err = utils.EncryptAES(data, []byte("short"))
	if err != utils.ErrInvalidKeySize {
		t.Errorf("EncryptAES() with invalid key size expected ErrInvalidKeySize, got %v", err)
	}
}

func TestDecryptAES(t *testing.T) {
	key := []byte("1234567890123456") // 16 bytes
	originalData := []byte("test data")

	// First encrypt
	encrypted, err := utils.EncryptAES(originalData, key)
	if err != nil {
		t.Errorf("EncryptAES() unexpected error: %v", err)
	}

	// Then decrypt
	decrypted, err := utils.DecryptAES(encrypted, key)
	if err != nil {
		t.Errorf("DecryptAES() unexpected error: %v", err)
	}

	if string(decrypted) != string(originalData) {
		t.Errorf("DecryptAES() = %s, want %s", string(decrypted), string(originalData))
	}

	// Test with empty encrypted data
	_, err = utils.DecryptAES("", key)
	if err != utils.ErrInvalidInput {
		t.Errorf("DecryptAES() with empty data expected ErrInvalidInput, got %v", err)
	}

	// Test with empty key
	_, err = utils.DecryptAES(encrypted, []byte{})
	if err != utils.ErrInvalidInput {
		t.Errorf("DecryptAES() with empty key expected ErrInvalidInput, got %v", err)
	}

	// Test with invalid key size
	_, err = utils.DecryptAES(encrypted, []byte("short"))
	if err != utils.ErrInvalidKeySize {
		t.Errorf("DecryptAES() with invalid key size expected ErrInvalidKeySize, got %v", err)
	}

	// Test with invalid base64
	_, err = utils.DecryptAES("invalid-base64", key)
	if err == nil {
		t.Error("DecryptAES() with invalid base64 expected error")
	}
}

func TestGenerateAESKey(t *testing.T) {
	// Test valid key sizes
	sizes := []int{16, 24, 32}
	for _, size := range sizes {
		key, err := utils.GenerateAESKey(size)
		if err != nil {
			t.Errorf("GenerateAESKey(%d) unexpected error: %v", size, err)
		}
		if len(key) != size {
			t.Errorf("GenerateAESKey(%d) returned key of length %d, want %d", size, len(key), size)
		}
	}

	// Test invalid key size
	_, err := utils.GenerateAESKey(8)
	if err != utils.ErrInvalidKeySize {
		t.Errorf("GenerateAESKey(8) expected ErrInvalidKeySize, got %v", err)
	}
}

func TestHMACSHA256(t *testing.T) {
	data := []byte("test data")
	key := []byte("test key")

	signature := utils.HMACSHA256(data, key)
	if signature == "" {
		t.Error("HMACSHA256() returned empty signature")
	}

	// Test with empty data
	emptySignature := utils.HMACSHA256([]byte{}, key)
	if emptySignature != "" {
		t.Error("HMACSHA256() with empty data should return empty string")
	}

	// Test with empty key
	emptySignature = utils.HMACSHA256(data, []byte{})
	if emptySignature != "" {
		t.Error("HMACSHA256() with empty key should return empty string")
	}
}

func TestVerifyHMACSHA256(t *testing.T) {
	data := []byte("test data")
	key := []byte("test key")

	// Generate signature
	signature := utils.HMACSHA256(data, key)
	if signature == "" {
		t.Fatal("HMACSHA256() returned empty signature")
	}

	// Verify valid signature
	if !utils.VerifyHMACSHA256(data, signature, key) {
		t.Error("VerifyHMACSHA256() should return true for valid signature")
	}

	// Verify with wrong data
	if utils.VerifyHMACSHA256([]byte("wrong data"), signature, key) {
		t.Error("VerifyHMACSHA256() should return false for wrong data")
	}

	// Verify with wrong key
	if utils.VerifyHMACSHA256(data, signature, []byte("wrong key")) {
		t.Error("VerifyHMACSHA256() should return false for wrong key")
	}

	// Verify with empty data
	if utils.VerifyHMACSHA256([]byte{}, signature, key) {
		t.Error("VerifyHMACSHA256() should return false for empty data")
	}

	// Verify with empty signature
	if utils.VerifyHMACSHA256(data, "", key) {
		t.Error("VerifyHMACSHA256() should return false for empty signature")
	}

	// Verify with empty key
	if utils.VerifyHMACSHA256(data, signature, []byte{}) {
		t.Error("VerifyHMACSHA256() should return false for empty key")
	}

	// Verify with invalid base64 signature
	if utils.VerifyHMACSHA256(data, "invalid-base64", key) {
		t.Error("VerifyHMACSHA256() should return false for invalid base64 signature")
	}
}

func TestGenerateSalt(t *testing.T) {
	// Test valid length
	salt, err := utils.GenerateSalt(16)
	if err != nil {
		t.Errorf("GenerateSalt(16) unexpected error: %v", err)
	}
	if len(salt) != 16 {
		t.Errorf("GenerateSalt(16) returned salt of length %d, want 16", len(salt))
	}

	// Test invalid length
	_, err = utils.GenerateSalt(0)
	if err != utils.ErrInvalidInput {
		t.Errorf("GenerateSalt(0) expected ErrInvalidInput, got %v", err)
	}

	_, err = utils.GenerateSalt(-1)
	if err != utils.ErrInvalidInput {
		t.Errorf("GenerateSalt(-1) expected ErrInvalidInput, got %v", err)
	}
}

func TestSecureCompare(t *testing.T) {
	a := []byte("test data")
	b := []byte("test data")
	c := []byte("different data")

	// Test equal slices
	if !utils.SecureCompare(a, b) {
		t.Error("SecureCompare() should return true for equal slices")
	}

	// Test different slices
	if utils.SecureCompare(a, c) {
		t.Error("SecureCompare() should return false for different slices")
	}

	// Test empty slices
	if !utils.SecureCompare([]byte{}, []byte{}) {
		t.Error("SecureCompare() should return true for empty slices")
	}

	// Test different length slices
	if utils.SecureCompare([]byte("short"), []byte("longer")) {
		t.Error("SecureCompare() should return false for different length slices")
	}
}
