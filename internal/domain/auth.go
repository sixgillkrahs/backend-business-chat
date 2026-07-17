package domain

import "time"

type Action struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Code        string    `json:"code" db:"code"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Resource struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Code        string    `json:"code" db:"code"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

var DefaultResources = []Resource{
	{
		Name:        "Quản lý người dùng",
		Code:        "USER_MANAGEMENT",
		Description: "Chức năng xem, tạo, sửa, xóa người dùng",
	},
	{
		Name:        "Quản lý phân quyền",
		Code:        "ROLE_MANAGEMENT",
		Description: "Chức năng cấu hình vai trò và quyền hạn",
	},
	{
		Name:        "Cấu hình hệ thống",
		Code:        "SYSTEM_SETTINGS",
		Description: "Chức năng điều chỉnh các thông số hệ thống",
	},
}

type Policy struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	ActionID    int       `json:"action_id" db:"action_id"`
	ResourceID  int       `json:"resource_id" db:"resource_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
