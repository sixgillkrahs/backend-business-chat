package utils

import "sync"

// Định nghĩa các Key đa ngôn ngữ chuẩn (Constants)
const (
	KeySuccess             = "msg.success"
	KeyCreatedSuccess      = "msg.created_success"
	KeyDeletedSuccess      = "msg.deleted_success"
	KeyValidationFailed    = "msg.validation_failed"
	KeyInternalServerError = "msg.internal_server_error"

	KeyErrInvalidCredentials = "err.invalid_credentials"
	KeyErrUserNotFound       = "err.user_not_found"
	KeyErrUserLocked         = "err.user_locked"
	KeyErrUsernameTooShort   = "err.username_too_short"
	KeyErrInvalidEmail       = "err.invalid_email"
	KeyErrPermissionNotFound = "err.permission_not_found"
	KeyErrInvalidID          = "err.invalid_id"
)

var (
	translationMu sync.RWMutex
	translations  = map[string]map[string]string{
		"vi": {
			KeySuccess:             "Thành công",
			KeyCreatedSuccess:      "Tạo thành công",
			KeyDeletedSuccess:      "Xóa thành công",
			KeyValidationFailed:    "Dữ liệu không hợp lệ",
			KeyInternalServerError: "Lỗi hệ thống nội bộ",

			KeyErrInvalidCredentials: "Tài khoản hoặc mật khẩu không chính xác",
			KeyErrUserNotFound:       "Người dùng không tồn tại",
			KeyErrUserLocked:         "Người dùng đã bị khóa",
			KeyErrUsernameTooShort:   "Username phải có ít nhất 3 ký tự",
			KeyErrInvalidEmail:       "Email không hợp lệ",
			KeyErrPermissionNotFound: "Permission không tồn tại trong hệ thống",
			KeyErrInvalidID:          "ID không hợp lệ",

			"Success":               "Thành công",
			"Created successfully":  "Tạo thành công",
			"Tạo thành công":        "Tạo thành công",
			"Xóa thành công":        "Xóa thành công",
			"Validation failed":     "Dữ liệu không hợp lệ",
			"Internal server error": "Lỗi hệ thống nội bộ",
			"Lỗi hệ thống nội bộ":   "Lỗi hệ thống nội bộ",
			"tài khoản hoặc mật khẩu không chính xác": "Tài khoản hoặc mật khẩu không chính xác",
			"người dùng không tồn tại":                "Người dùng không tồn tại",
			"Người dùng đã bị khóa":                   "Người dùng đã bị khóa",
			"username phải có ít nhất 3 ký tự":        "Username phải có ít nhất 3 ký tự",
			"email không hợp lệ":                      "Email không hợp lệ",
			"permission không tồn tại trong hệ thống": "Permission không tồn tại trong hệ thống",
			"ID không hợp lệ":                         "ID không hợp lệ",
			"tên permission đã tồn tại trong hệ thống": "Tên permission đã tồn tại trong hệ thống",
			"một hoặc nhiều policy không tồn tại trong hệ thống": "Một hoặc nhiều policy không tồn tại trong hệ thống",
			"permission phải chứa ít nhất một policy": "Permission phải chứa ít nhất một policy",
			"tên permission phải từ 3 đến 50 ký tự":   "Tên permission phải từ 3 đến 50 ký tự",
			"mô tả không được vượt quá 255 ký tự":     "Mô tả không được vượt quá 255 ký tự",
			"Tạo role thành công":                     "Tạo role thành công",
			"tên vai trò đã tồn tại trong hệ thống":   "Tên vai trò đã tồn tại trong hệ thống",
			"vai trò không tồn tại trong hệ thống":    "Vai trò không tồn tại trong hệ thống",
			"vai trò phải chứa ít nhất một permission": "Vai trò phải chứa ít nhất một permission",
			"tên vai trò phải từ 3 đến 50 ký tự":     "Tên vai trò phải từ 3 đến 50 ký tự",
		},
		"en": {
			KeySuccess:             "Success",
			KeyCreatedSuccess:      "Created successfully",
			KeyDeletedSuccess:      "Deleted successfully",
			KeyValidationFailed:    "Validation failed",
			KeyInternalServerError: "Internal server error",

			KeyErrInvalidCredentials: "Invalid username or password",
			KeyErrUserNotFound:       "User not found",
			KeyErrUserLocked:         "User account is locked",
			KeyErrUsernameTooShort:   "Username must be at least 3 characters",
			KeyErrInvalidEmail:       "Invalid email address",
			KeyErrPermissionNotFound: "Permission not found",
			KeyErrInvalidID:          "Invalid ID",

			"Success":               "Success",
			"Created successfully":  "Created successfully",
			"Tạo thành công":        "Created successfully",
			"Xóa thành công":        "Deleted successfully",
			"Validation failed":     "Validation failed",
			"Internal server error": "Internal server error",
			"Lỗi hệ thống nội bộ":   "Internal server error",
			"tài khoản hoặc mật khẩu không chính xác": "Invalid username or password",
			"người dùng không tồn tại":                "User not found",
			"Người dùng đã bị khóa":                   "User account is locked",
			"username phải có ít nhất 3 ký tự":        "Username must be at least 3 characters",
			"email không hợp lệ":                      "Invalid email address",
			"permission không tồn tại trong hệ thống": "Permission not found",
			"ID không hợp lệ":                         "Invalid ID",
			"tên permission đã tồn tại trong hệ thống": "Permission name already exists",
			"một hoặc nhiều policy không tồn tại trong hệ thống": "One or more policies do not exist",
			"permission phải chứa ít nhất một policy": "Permission must contain at least one policy",
			"tên permission phải từ 3 đến 50 ký tự":   "Permission name must be between 3 and 50 characters",
			"mô tả không được vượt quá 255 ký tự":     "Description must not exceed 255 characters",
			"Tạo role thành công":                     "Role created successfully",
			"tên vai trò đã tồn tại trong hệ thống":   "Role name already exists",
			"vai trò không tồn tại trong hệ thống":    "Role not found",
			"vai trò phải chứa ít nhất một permission": "Role must contain at least one permission",
			"tên vai trò phải từ 3 đến 50 ký tự":     "Role name must be between 3 and 50 characters",
		},
	}
)
