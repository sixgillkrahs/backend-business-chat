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

type Auth struct {
	ID           string    `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"password_hash" db:"password_hash"`
	RoleId       int       `json:"role_id" db:"role_id"`
	UserId       int       `json:"user_id" db:"user_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy    string    `json:"created_by" db:"created_by"`
	UpdatedBy    string    `json:"updated_by" db:"updated_by"`
	DeletedBy    string    `json:"deleted_by" db:"deleted_by"`
	DeletedAt    time.Time `json:"deleted_at" db:"deleted_at"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	IsDeleted    bool      `json:"is_deleted" db:"is_deleted"`
}
