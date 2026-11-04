package utils_test

import (
	"testing"

	"github.com/sixgillkrahs/backend-business-chat/pkg/utils"
)

func TestIsPasswordComplex(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		expected bool
	}{
		{"Valid password with symbols", "Password123!", true},
		{"Valid password with @ and special", "123456aA@1", true},
		{"Valid password with underscore", "Secure_Pass_1", true},
		{"Too short (< 8)", "Pass1!", false},
		{"Too long (> 72)", "A1!" + string(make([]byte, 71)), false},
		{"No uppercase", "password123!", false},
		{"No lowercase", "PASSWORD123!", false},
		{"No digit", "Password!!!!", false},
		{"No special char", "Password1234", false},
		{"Empty password", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.IsPasswordComplex(tc.password)
			if result != tc.expected {
				t.Errorf("IsPasswordComplex(%q) = %v, want %v", tc.password, result, tc.expected)
			}
		})
	}
}
