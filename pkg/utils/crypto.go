package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

var (
	ErrInvalidKeySize = errors.New("invalid key size: must be 16, 24, or 32 bytes")
	ErrInvalidInput   = errors.New("invalid input parameters")
	ErrCipherTooShort = errors.New("ciphertext too short")
)

// EncryptAES encrypts data using AES-GCM
// data: the plaintext to encrypt
// key: the encryption key (must be 16, 24, or 32 bytes)
// Returns base64-encoded encrypted data or error
func EncryptAES(data []byte, key []byte) (string, error) {
	if len(data) == 0 || len(key) == 0 {
		return "", ErrInvalidInput
	}

	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", ErrInvalidKeySize
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES decrypts data using AES-GCM
// encryptedData: base64-encoded encrypted data
// key: the decryption key (must be 16, 24, or 32 bytes)
// Returns decrypted plaintext or error
func DecryptAES(encryptedData string, key []byte) ([]byte, error) {
	if encryptedData == "" || len(key) == 0 {
		return nil, ErrInvalidInput
	}

	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, ErrInvalidKeySize
	}

	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, ErrCipherTooShort
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// GenerateAESKey generates a random AES key
// size: key size in bytes (must be 16, 24, or 32)
// Returns random key or error
func GenerateAESKey(size int) ([]byte, error) {
	if size != 16 && size != 24 && size != 32 {
		return nil, ErrInvalidKeySize
	}

	key := make([]byte, size)
	_, err := rand.Read(key)
	return key, err
}

// HMACSHA256 creates HMAC-SHA256 signature
// data: the data to sign
// key: the signing key
// Returns base64-encoded signature
func HMACSHA256(data []byte, key []byte) string {
	if len(data) == 0 || len(key) == 0 {
		return ""
	}

	h := hmac.New(sha256.New, key)
	h.Write(data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// VerifyHMACSHA256 verifies HMAC-SHA256 signature using constant-time comparison
// data: the original data
// signature: base64-encoded signature to verify
// key: the signing key
// Returns true if signature is valid, false otherwise
func VerifyHMACSHA256(data []byte, signature string, key []byte) bool {
	if len(data) == 0 || signature == "" || len(key) == 0 {
		return false
	}

	expectedSignature := HMACSHA256(data, key)
	if expectedSignature == "" {
		return false
	}

	// Decode both signatures for comparison
	expectedBytes, err1 := base64.StdEncoding.DecodeString(expectedSignature)
	actualBytes, err2 := base64.StdEncoding.DecodeString(signature)

	if err1 != nil || err2 != nil {
		return false
	}

	return hmac.Equal(expectedBytes, actualBytes)
}

// GenerateSalt generates a random salt
// length: salt length in bytes
// Returns random salt or error
func GenerateSalt(length int) ([]byte, error) {
	if length <= 0 {
		return nil, ErrInvalidInput
	}

	salt := make([]byte, length)
	_, err := rand.Read(salt)
	return salt, err
}

// SecureCompare performs constant-time comparison of byte slices
// a, b: byte slices to compare
// Returns true if slices are equal, false otherwise
func SecureCompare(a, b []byte) bool {
	return hmac.Equal(a, b)
}
