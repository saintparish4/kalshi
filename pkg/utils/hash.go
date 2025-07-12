package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"time"
)

// HashString returns SHA256 hash of a string
func HashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// HashStringSHA1 returns SHA1 hash of a string
// Deprecated: Use HashString instead for better security
func HashStringSHA1(s string) string {
	hash := sha1.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

// HashStringMD5 returns MD5 hash of a string
// Deprecated: Use HashString instead for better security
func HashStringMD5(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

// HashBytes returns SHA256 hash of byte slice
func HashBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// GenerateAPIKey generates a secure API key
func GenerateAPIKey() string {
	data := fmt.Sprintf("%d-%s", time.Now().UnixNano(), GenerateRandomString(32))
	return HashString(data)
}

// GenerateShortHash generates a short hash for IDs
// Note: 8-character hashes have collision risk, use for non-critical IDs only
func GenerateShortHash(input string) string {
	return HashString(input)[:8]
}

// HashWithSalt adds salt to hashing for security
func HashWithSalt(data, salt string) string {
	combined := data + salt
	return HashString(combined)
}

// VerifyHash verifies if plain text matches hash with salt
func VerifyHash(plaintext, salt, hashedValue string) bool {
	return HashWithSalt(plaintext, salt) == hashedValue
}

// Hash interface for different hash algorithms
type Hasher interface {
	Hash(data []byte) string
	HashString(s string) string
}

// SHA256Hasher implements SHA256 hashing
type SHA256Hasher struct{}

func (h SHA256Hasher) Hash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (h SHA256Hasher) HashString(s string) string {
	return h.Hash([]byte(s))
}

// SHA512Hasher implements SHA512 hashing
type SHA512Hasher struct{}

func (h SHA512Hasher) Hash(data []byte) string {
	hash := sha512.Sum512(data)
	return hex.EncodeToString(hash[:])
}

func (h SHA512Hasher) HashString(s string) string {
	return h.Hash([]byte(s))
}

// NewHasher creates a hasher based on algorithm name
func NewHasher(algorithm string) Hasher {
	switch algorithm {
	case "sha256":
		return SHA256Hasher{}
	case "sha512":
		return SHA512Hasher{}
	default:
		return SHA256Hasher{}
	}
}
