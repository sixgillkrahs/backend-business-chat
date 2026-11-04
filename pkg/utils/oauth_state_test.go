package utils_test

import (
	"testing"
	"time"

	"github.com/sixgillkrahs/backend-business-chat/pkg/utils"
)

func TestOAuthState_GenerateAndValidate(t *testing.T) {
	secretKey := "test-secret-key-12345"

	state, err := utils.GenerateOAuthState(secretKey)
	if err != nil {
		t.Fatalf("unexpected error generating state: %v", err)
	}
	if state == "" {
		t.Fatalf("expected non-empty state")
	}

	// Valid state should pass
	if !utils.ValidateOAuthState(state, secretKey, 5*time.Minute) {
		t.Errorf("expected valid state to pass validation")
	}

	// Wrong secret key should fail
	if utils.ValidateOAuthState(state, "wrong-secret-key", 5*time.Minute) {
		t.Errorf("expected wrong secret key to fail validation")
	}

	// Tampered state should fail
	tampered := state + "tamper"
	if utils.ValidateOAuthState(tampered, secretKey, 5*time.Minute) {
		t.Errorf("expected tampered state to fail validation")
	}

	// Empty state or empty key should fail
	if utils.ValidateOAuthState("", secretKey, 5*time.Minute) {
		t.Errorf("expected empty state to fail validation")
	}
	if utils.ValidateOAuthState(state, "", 5*time.Minute) {
		t.Errorf("expected empty key to fail validation")
	}

	// Expired state (maxAge = 0 seconds)
	time.Sleep(1100 * time.Millisecond)
	if utils.ValidateOAuthState(state, secretKey, 1*time.Second) {
		t.Errorf("expected expired state to fail validation")
	}
}
