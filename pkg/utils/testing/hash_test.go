package testing

import (
	"testing"

	"kalshi/pkg/utils"
)

func TestHashString(t *testing.T) {
	input := "test string"
	hash := utils.HashString(input)

	if hash == "" {
		t.Error("HashString() returned empty hash")
	}

	// Hash should be deterministic
	hash2 := utils.HashString(input)
	if hash != hash2 {
		t.Error("HashString() should return same hash for same input")
	}

	// Different inputs should produce different hashes
	hash3 := utils.HashString("different string")
	if hash == hash3 {
		t.Error("HashString() should return different hashes for different inputs")
	}
}

func TestHashStringSHA1(t *testing.T) {
	input := "test string"
	hash := utils.HashStringSHA1(input)

	if hash == "" {
		t.Error("HashStringSHA1() returned empty hash")
	}

	// Hash should be deterministic
	hash2 := utils.HashStringSHA1(input)
	if hash != hash2 {
		t.Error("HashStringSHA1() should return same hash for same input")
	}
}

func TestHashStringMD5(t *testing.T) {
	input := "test string"
	hash := utils.HashStringMD5(input)

	if hash == "" {
		t.Error("HashStringMD5() returned empty hash")
	}

	// Hash should be deterministic
	hash2 := utils.HashStringMD5(input)
	if hash != hash2 {
		t.Error("HashStringMD5() should return same hash for same input")
	}
}

func TestHashBytes(t *testing.T) {
	input := []byte("test data")
	hash := utils.HashBytes(input)

	if hash == "" {
		t.Error("HashBytes() returned empty hash")
	}

	// Hash should be deterministic
	hash2 := utils.HashBytes(input)
	if hash != hash2 {
		t.Error("HashBytes() should return same hash for same input")
	}

	// Different inputs should produce different hashes
	hash3 := utils.HashBytes([]byte("different data"))
	if hash == hash3 {
		t.Error("HashBytes() should return different hashes for different inputs")
	}
}

func TestGenerateAPIKey(t *testing.T) {
	key1 := utils.GenerateAPIKey()
	key2 := utils.GenerateAPIKey()

	if key1 == "" {
		t.Error("GenerateAPIKey() returned empty key")
	}

	if key1 == key2 {
		t.Error("GenerateAPIKey() should return different keys")
	}
}

func TestGenerateShortHash(t *testing.T) {
	input := "test input"
	hash := utils.GenerateShortHash(input)

	if len(hash) != 8 {
		t.Errorf("GenerateShortHash() returned hash of length %d, want 8", len(hash))
	}

	// Hash should be deterministic
	hash2 := utils.GenerateShortHash(input)
	if hash != hash2 {
		t.Error("GenerateShortHash() should return same hash for same input")
	}
}

func TestHashWithSalt(t *testing.T) {
	data := "test data"
	salt := "test salt"
	hash := utils.HashWithSalt(data, salt)

	if hash == "" {
		t.Error("HashWithSalt() returned empty hash")
	}

	// Same data and salt should produce same hash
	hash2 := utils.HashWithSalt(data, salt)
	if hash != hash2 {
		t.Error("HashWithSalt() should return same hash for same input")
	}

	// Different salt should produce different hash
	hash3 := utils.HashWithSalt(data, "different salt")
	if hash == hash3 {
		t.Error("HashWithSalt() should return different hash for different salt")
	}
}

func TestVerifyHash(t *testing.T) {
	plaintext := "test password"
	salt := "test salt"
	hashedValue := utils.HashWithSalt(plaintext, salt)

	// Test valid verification
	if !utils.VerifyHash(plaintext, salt, hashedValue) {
		t.Error("VerifyHash() should return true for valid hash")
	}

	// Test wrong plaintext
	if utils.VerifyHash("wrong password", salt, hashedValue) {
		t.Error("VerifyHash() should return false for wrong plaintext")
	}

	// Test wrong salt
	if utils.VerifyHash(plaintext, "wrong salt", hashedValue) {
		t.Error("VerifyHash() should return false for wrong salt")
	}

	// Test wrong hash
	if utils.VerifyHash(plaintext, salt, "wrong hash") {
		t.Error("VerifyHash() should return false for wrong hash")
	}
}

func TestSHA256Hasher(t *testing.T) {
	hasher := utils.SHA256Hasher{}
	input := "test data"

	hash := hasher.Hash([]byte(input))
	if hash == "" {
		t.Error("SHA256Hasher.Hash() returned empty hash")
	}

	hashString := hasher.HashString(input)
	if hashString == "" {
		t.Error("SHA256Hasher.HashString() returned empty hash")
	}

	// Both methods should produce same result
	if hash != hashString {
		t.Error("SHA256Hasher.Hash() and HashString() should return same result")
	}
}

func TestSHA512Hasher(t *testing.T) {
	hasher := utils.SHA512Hasher{}
	input := "test data"

	hash := hasher.Hash([]byte(input))
	if hash == "" {
		t.Error("SHA512Hasher.Hash() returned empty hash")
	}

	hashString := hasher.HashString(input)
	if hashString == "" {
		t.Error("SHA512Hasher.HashString() returned empty hash")
	}

	// Both methods should produce same result
	if hash != hashString {
		t.Error("SHA512Hasher.Hash() and HashString() should return same result")
	}
}

func TestNewHasher(t *testing.T) {
	// Test SHA256 hasher
	hasher := utils.NewHasher("sha256")
	if _, ok := hasher.(utils.SHA256Hasher); !ok {
		t.Error("NewHasher('sha256') should return SHA256Hasher")
	}

	// Test SHA512 hasher
	hasher = utils.NewHasher("sha512")
	if _, ok := hasher.(utils.SHA512Hasher); !ok {
		t.Error("NewHasher('sha512') should return SHA512Hasher")
	}

	// Test default case
	hasher = utils.NewHasher("unknown")
	if _, ok := hasher.(utils.SHA256Hasher); !ok {
		t.Error("NewHasher('unknown') should return SHA256Hasher as default")
	}
}
