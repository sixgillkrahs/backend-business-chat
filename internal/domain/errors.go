package domain

import "errors"

var (
	ErrInvalidCredentials      = errors.New("tài khoản hoặc mật khẩu không chính xác")
	ErrUserNotFound            = errors.New("người dùng không tồn tại")
	ErrUserLocked              = errors.New("Người dùng đã bị khóa")
	ErrPermissionNameDuplicate = errors.New("tên permission đã tồn tại trong hệ thống")
	ErrPolicyNotFound          = errors.New("một hoặc nhiều policy không tồn tại trong hệ thống")
	ErrEmptyPolicies           = errors.New("permission phải chứa ít nhất một policy")
)
