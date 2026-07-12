package domain

import (
	"errors"
	"time"
)

type UserStatus string

const (
	UserStatusPending  UserStatus = "pending"
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
)

type UserRole string

const (
	UserRoleAdmin  UserRole = "admin"
	UserRoleMember UserRole = "member"
)

type User struct {
	ID           int64      `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Status       UserStatus `json:"status"`
	Role         UserRole   `json:"role"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

var (
	ErrUsernameTooShort = errors.New("username phải có ít nhất 3 ký tự")
	ErrInvalidEmail     = errors.New("email không hợp lệ")
)

func NewUser(username, email, passwordHash string) (*User, error) {
	if len(username) < 3 {
		return nil, ErrUsernameTooShort
	}
	if email == "" {
		return nil, ErrInvalidEmail
	}

	return &User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Status:       UserStatusPending,
		Role:         UserRoleMember,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}
