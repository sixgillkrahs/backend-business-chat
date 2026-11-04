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
	ErrSystemRoleProtected        = errors.New("không thể chỉnh sửa vai trò hệ thống")
	ErrSystemRoleCannotDelete     = errors.New("không thể xóa vai trò mặc định của hệ thống")
	ErrSystemRoleCannotDeactivate = errors.New("không thể hủy kích hoạt vai trò mặc định của hệ thống")
	ErrRoleInUse                  = errors.New("không thể xóa vai trò đang được gán cho người dùng")
	ErrUsernameDuplicate       = errors.New("tên đăng nhập đã tồn tại trong hệ thống")
	ErrEmailDuplicate          = errors.New("địa chỉ email đã tồn tại trong hệ thống")
	ErrPasswordTooShort           = errors.New("mật khẩu phải có độ dài từ 8 đến 72 ký tự")
	ErrSystemUserCannotDelete     = errors.New("không thể xóa tài khoản quản trị hệ thống")
	ErrCannotDeleteSelf           = errors.New("không thể tự xóa tài khoản của chính mình")
	ErrRoleLocked                 = errors.New("vai trò của người dùng đã bị vô hiệu hóa")
	ErrInvalidRefreshToken        = errors.New("refresh token không hợp lệ hoặc đã hết hạn")
)
