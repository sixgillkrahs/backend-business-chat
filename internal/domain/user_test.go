package domain_test

import (
	"strings"
	"testing"

	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
)

func TestNewUser_Valid(t *testing.T) {
	user, err := domain.NewUser("john_doe", "john@example.com", "John Doe", "admin")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if user.Username != "john_doe" {
		t.Errorf("expected username john_doe, got: %s", user.Username)
	}
	if user.Email != "john@example.com" {
		t.Errorf("expected email john@example.com, got: %s", user.Email)
	}
	if user.FullName != "John Doe" {
		t.Errorf("expected fullName John Doe, got: %s", user.FullName)
	}
	if !user.IsActive || user.IsDeleted {
		t.Errorf("expected user to be active and not deleted")
	}
	if !user.IsActiveUser() {
		t.Errorf("expected IsActiveUser to be true")
	}
}

func TestNewUser_InvalidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
	}{
		{"Too short", "ab"},
		{"Too long", strings.Repeat("a", 51)},
		{"Contains special chars", "john@doe!"},
		{"Contains spaces", "john doe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewUser(tt.username, "john@example.com", "John Doe", "admin")
			if err != domain.ErrUsernameInvalid {
				t.Fatalf("expected ErrUsernameInvalid, got: %v", err)
			}
		})
	}
}

func TestNewUser_InvalidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{"Empty email", ""},
		{"Missing @", "johnexample.com"},
		{"Missing domain", "john@"},
		{"Missing TLD", "john@domain"},
		{"Too long", strings.Repeat("a", 100) + "@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewUser("john_doe", tt.email, "John Doe", "admin")
			if err != domain.ErrEmailInvalid {
				t.Fatalf("expected ErrEmailInvalid, got: %v", err)
			}
		})
	}
}

func TestNewUser_InvalidFullName(t *testing.T) {
	tests := []struct {
		name     string
		fullName string
	}{
		{"Empty", ""},
		{"Only whitespace", "   "},
		{"Too long", strings.Repeat("a", 101)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewUser("john_doe", "john@example.com", tt.fullName, "admin")
			if err != domain.ErrFullNameInvalid {
				t.Fatalf("expected ErrFullNameInvalid, got: %v", err)
			}
		})
	}
}

func TestUser_Lifecycle(t *testing.T) {
	user, err := domain.NewUser("john_doe", "john@example.com", "John Doe", "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user.Deactivate("supervisor")
	if user.IsActive {
		t.Errorf("expected user to be inactive")
	}
	if user.IsActiveUser() {
		t.Errorf("expected IsActiveUser to be false")
	}
	if user.UpdatedBy != "supervisor" {
		t.Errorf("expected updatedBy supervisor, got: %s", user.UpdatedBy)
	}

	user.Activate("supervisor")
	if !user.IsActive {
		t.Errorf("expected user to be active")
	}

	user.SoftDelete("admin")
	if !user.IsDeleted {
		t.Errorf("expected user to be deleted")
	}
	if user.IsActiveUser() {
		t.Errorf("expected IsActiveUser to be false when deleted")
	}
	if user.DeletedBy == nil || *user.DeletedBy != "admin" {
		t.Errorf("expected deletedBy admin")
	}
}
