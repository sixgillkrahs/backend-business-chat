package domain

import "errors"

var (
	ErrInvalidCredentials      = errors.New("tài khoản hoặc mật khẩu không chính xác")
	ErrUserNotFound            = errors.New("người dùng không tồn tại")
	ErrUserLocked              = errors.New("Người dùng đã bị khóa")
	ErrPermissionNameDuplicate = errors.New("tên permission đã tồn tại trong hệ thống")
	ErrPolicyNotFound          = errors.New("một hoặc nhiều policy không tồn tại trong hệ thống")
	ErrEmptyPolicies           = errors.New("permission phải chứa ít nhất một policy")
	ErrPermissionNotFound      = errors.New("permission không tồn tại trong hệ thống")
	ErrRoleNameDuplicate       = errors.New("tên vai trò đã tồn tại trong hệ thống")
	ErrRoleNotFound            = errors.New("vai trò không tồn tại trong hệ thống")
	ErrEmptyRolePermissions    = errors.New("vai trò phải chứa ít nhất một permission")
	ErrSystemRoleProtected     = errors.New("không thể chỉnh sửa vai trò hệ thống")
)
