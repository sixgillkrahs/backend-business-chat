package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrUsernameInvalid = errors.New("tên đăng nhập phải từ 3 đến 50 ký tự và không chứa ký tự đặc biệt")
	ErrEmailInvalid    = errors.New("địa chỉ email không hợp lệ")
	ErrFullNameInvalid = errors.New("họ và tên không được để trống và tối đa 100 ký tự")
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_.\-]+$`)
)

type User struct {
	ID          int        `json:"id" db:"id"`
	Username    string     `json:"username" db:"username"`
	Email       string     `json:"email" db:"email"`
	FullName    string     `json:"full_name" db:"full_name"`
	AvatarURL   string     `json:"avatar_url,omitempty" db:"avatar_url"`
	PhoneNumber string     `json:"phone_number,omitempty" db:"phone_number"`
	Address     string     `json:"address,omitempty" db:"address"`
	RoleID      *int       `json:"role_id,omitempty" db:"role_id"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	IsDeleted   bool       `json:"is_deleted" db:"is_deleted"`
	CreatedBy   string     `json:"created_by" db:"created_by"`
	UpdatedBy   string     `json:"updated_by" db:"updated_by"`
	DeletedBy   *string    `json:"deleted_by,omitempty" db:"deleted_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	Role *Role `json:"role,omitempty" db:"-"`
	Auth *Auth `json:"auth,omitempty" db:"-"`
}

func (User) TableName() string {
	return "users"
}

func NewUser(username, email, fullName, createdBy string) (*User, error) {
	u := &User{
		Username:  username,
		Email:     email,
		FullName:  fullName,
		IsActive:  true,
		IsDeleted: false,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := u.Validate(); err != nil {
		return nil, err
	}

	return u, nil
}

func (u *User) Validate() error {
	u.Username = strings.TrimSpace(u.Username)
	if len(u.Username) < 3 || len(u.Username) > 50 || !usernameRegex.MatchString(u.Username) {
		return ErrUsernameInvalid
	}

	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	if len(u.Email) == 0 || len(u.Email) > 100 || !emailRegex.MatchString(u.Email) {
		return ErrEmailInvalid
	}

	u.FullName = strings.TrimSpace(u.FullName)
	if len(u.FullName) == 0 || len(u.FullName) > 100 {
		return ErrFullNameInvalid
	}

	return nil
}

func (u *User) IsActiveUser() bool {
	return u.IsActive && !u.IsDeleted
}

func (u *User) Deactivate(by string) {
	now := time.Now()
	u.IsActive = false
	u.UpdatedBy = by
	u.UpdatedAt = now
}

func (u *User) Activate(by string) {
	now := time.Now()
	u.IsActive = true
	u.UpdatedBy = by
	u.UpdatedAt = now
}

func (u *User) SoftDelete(by string) {
	now := time.Now()
	u.IsDeleted = true
	u.DeletedBy = &by
	u.DeletedAt = &now
	u.UpdatedBy = by
	u.UpdatedAt = now
}
