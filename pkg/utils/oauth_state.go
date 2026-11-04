package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GenerateOAuthState creates a cryptographically signed state parameter for CSRF mitigation.
// Structure: base64url(nonce + "." + timestamp + "." + hmacSignature)
func GenerateOAuthState(secretKey string) (string, error) {
	if secretKey == "" {
		return "", errors.New("secretKey cannot be empty")
	}

	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", fmt.Errorf("failed to generate random nonce: %w", err)
	}
	nonce := hex.EncodeToString(nonceBytes)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	payload := nonce + "." + timestamp
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	rawState := payload + "." + signature
	return base64.RawURLEncoding.EncodeToString([]byte(rawState)), nil
}

// ValidateOAuthState verifies that the OAuth state parameter is authentic, untampered, and unexpired.
func ValidateOAuthState(stateStr string, secretKey string, maxAge time.Duration) bool {
	if stateStr == "" || secretKey == "" {
		return false
	}

	decodedBytes, err := base64.RawURLEncoding.DecodeString(stateStr)
	if err != nil {
		return false
	}

	parts := strings.Split(string(decodedBytes), ".")
	if len(parts) != 3 {
		return false
	}

	nonce := parts[0]
	tsStr := parts[1]
	receivedSig := parts[2]

	if nonce == "" || tsStr == "" || receivedSig == "" {
		return false
	}

	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return false
	}

	stateTime := time.Unix(ts, 0)
	now := time.Now()
	if now.Before(stateTime) || now.After(stateTime.Add(maxAge)) {
		return false
	}

	payload := nonce + "." + tsStr
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(payload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(receivedSig), []byte(expectedSig))
}
