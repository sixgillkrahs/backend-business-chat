package utils

import "sync"

// Định nghĩa các Key đa ngôn ngữ chuẩn (Constants)
const (
	KeySuccess             = "msg.success"
	KeyCreatedSuccess      = "msg.created_success"
	KeyValidationFailed    = "msg.validation_failed"
	KeyInternalServerError = "msg.internal_server_error"

	KeyErrInvalidCredentials = "err.invalid_credentials"
	KeyErrUserNotFound       = "err.user_not_found"
	KeyErrUserLocked         = "err.user_locked"
	KeyErrUsernameTooShort   = "err.username_too_short"
	KeyErrInvalidEmail       = "err.invalid_email"
)

var (
	translationMu sync.RWMutex
	translations  = map[string]map[string]string{
		"vi": {
			KeySuccess:             "Thành công",
			KeyCreatedSuccess:      "Tạo thành công",
			KeyValidationFailed:    "Dữ liệu không hợp lệ",
			KeyInternalServerError: "Lỗi hệ thống nội bộ",

			KeyErrInvalidCredentials: "Tài khoản hoặc mật khẩu không chính xác",
			KeyErrUserNotFound:       "Người dùng không tồn tại",
			KeyErrUserLocked:         "Người dùng đã bị khóa",
			KeyErrUsernameTooShort:   "Username phải có ít nhất 3 ký tự",
			KeyErrInvalidEmail:       "Email không hợp lệ",

			"Success":               "Thành công",
			"Created successfully":  "Tạo thành công",
			"Tạo thành công":        "Tạo thành công",
			"Validation failed":     "Dữ liệu không hợp lệ",
			"Internal server error": "Lỗi hệ thống nội bộ",
			"Lỗi hệ thống nội bộ":   "Lỗi hệ thống nội bộ",
			"tài khoản hoặc mật khẩu không chính xác": "Tài khoản hoặc mật khẩu không chính xác",
			"người dùng không tồn tại":                "Người dùng không tồn tại",
			"Người dùng đã bị khóa":                   "Người dùng đã bị khóa",
			"username phải có ít nhất 3 ký tự":        "Username phải có ít nhất 3 ký tự",
			"email không hợp lệ":                      "Email không hợp lệ",
		},
		"en": {
			KeySuccess:             "Success",
			KeyCreatedSuccess:      "Created successfully",
			KeyValidationFailed:    "Validation failed",
			KeyInternalServerError: "Internal server error",

			KeyErrInvalidCredentials: "Invalid username or password",
			KeyErrUserNotFound:       "User not found",
			KeyErrUserLocked:         "User account is locked",
			KeyErrUsernameTooShort:   "Username must be at least 3 characters",
			KeyErrInvalidEmail:       "Invalid email address",

			"Success":               "Success",
			"Created successfully":  "Created successfully",
			"Tạo thành công":        "Created successfully",
			"Validation failed":     "Validation failed",
			"Internal server error": "Internal server error",
			"Lỗi hệ thống nội bộ":   "Internal server error",
			"tài khoản hoặc mật khẩu không chính xác": "Invalid username or password",
			"người dùng không tồn tại":                "User not found",
			"Người dùng đã bị khóa":                   "User account is locked",
			"username phải có ít nhất 3 ký tự":        "Username must be at least 3 characters",
			"email không hợp lệ":                      "Invalid email address",
		},
	}
)
