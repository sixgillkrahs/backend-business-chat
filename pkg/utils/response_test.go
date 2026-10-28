package utils

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sixgillkrahs/backend-business-chat/internal/interface/http/middleware"
)

func createTestContext(lang string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), middleware.LanguageKey, lang)
	c.Request = req.WithContext(ctx)
	return c, w
}

func TestResponse_Vietnamese(t *testing.T) {
	c, w := createTestContext("vi")

	Success(c, "data_test")

	if w.Code != http.StatusOK {
		t.Errorf("Expected code %d, got %d", http.StatusOK, w.Code)
	}

	var res Response
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if res.Message != "Thành công" {
		t.Errorf("Expected message 'Thành công', got '%s'", res.Message)
	}
}

func TestResponse_English(t *testing.T) {
	c, w := createTestContext("en")

	Success(c, "data_test")

	if w.Code != http.StatusOK {
		t.Errorf("Expected code %d, got %d", http.StatusOK, w.Code)
	}

	var res Response
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if res.Message != "Success" {
		t.Errorf("Expected message 'Success', got '%s'", res.Message)
	}
}

func TestHandleDomainError_i18n(t *testing.T) {
	errInvalid := errors.New("tài khoản hoặc mật khẩu không chính xác")
	domainErrMap := map[error]int{
		errInvalid: http.StatusUnauthorized,
	}

	// Case 1: English
	cEn, wEn := createTestContext("en")
	HandleDomainError(cEn, errInvalid, domainErrMap)

	var resEn Response
	_ = json.Unmarshal(wEn.Body.Bytes(), &resEn)
	if resEn.Message != "Invalid username or password" {
		t.Errorf("Expected 'Invalid username or password', got '%s'", resEn.Message)
	}

	// Case 2: Vietnamese
	cVi, wVi := createTestContext("vi")
	HandleDomainError(cVi, errInvalid, domainErrMap)

	var resVi Response
	_ = json.Unmarshal(wVi.Body.Bytes(), &resVi)
	if resVi.Message != "Tài khoản hoặc mật khẩu không chính xác" {
		t.Errorf("Expected 'Tài khoản hoặc mật khẩu không chính xác', got '%s'", resVi.Message)
	}
}
