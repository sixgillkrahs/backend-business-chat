package utils

import (
	"testing"
	"time"
)

func TestJWT_GenerateAndValidate(t *testing.T) {
	secret := "super-secret-key-12345"
	userID := int64(100)
	username := "admin"
	role := "admin"
	ip := "127.0.0.0"

	// 1. Generate Token
	tokenStr, err := GenerateToken(userID, username, role, ip, secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("Expected non-empty token string")
	}

	// 2. Validate Token
	claims, err := ValidateToken(tokenStr, secret)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}
	if claims.Username != username {
		t.Errorf("Expected Username '%s', got '%s'", username, claims.Username)
	}
	if claims.Role != role {
		t.Errorf("Expected Role '%s', got '%s'", role, claims.Role)
	}

	// 3. Test Invalid Secret Key
	_, err = ValidateToken(tokenStr, "wrong-secret-key")
	if err == nil {
		t.Error("Expected error when validating with wrong secret key, got nil")
	}
}
