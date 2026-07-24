package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("tài khoản hoặc mật khẩu không chính xác")
	ErrUserNotFound       = errors.New("người dùng không tồn tại")
	ErrUserLocked         = errors.New("Người dùng đã bị khóa")
)
