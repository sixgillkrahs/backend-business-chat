package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

var (
	ErrInvalidCiphertext = errors.New("ciphertext is invalid or corrupt")
	ErrEmptyPlaintext    = errors.New("plaintext cannot be empty")
)

// deriveKey derives a deterministic 32-byte key from secretKey using SHA-256 for AES-256.
func deriveKey(secretKey string) []byte {
	hash := sha256.Sum256([]byte(secretKey))
	return hash[:]
}

// EncryptAESGCM encrypts plaintext using AES-256-GCM with a cryptographically secure random nonce.
// Returns base64-encoded string containing: nonce (12 bytes) + ciphertext + authentication tag (16 bytes).
func EncryptAESGCM(plaintext string, secretKey string) (string, error) {
	if plaintext == "" {
		return "", ErrEmptyPlaintext
	}

	key := deriveKey(secretKey)
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

	// Seal appends ciphertext and authentication tag to nonce
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAESGCM decrypts a base64-encoded AES-256-GCM ciphertext and verifies its integrity.
func DecryptAESGCM(encryptedBase64 string, secretKey string) (string, error) {
	if encryptedBase64 == "" {
		return "", ErrInvalidCiphertext
	}

	data, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	key := deriveKey(secretKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	return string(plaintext), nil
}
