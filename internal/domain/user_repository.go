package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	CreateWithAuth(ctx context.Context, user *User, passwordHash string) error
	FindByID(ctx context.Context, id int) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int) error
	GetUsersPaging(ctx context.Context, search string, offset, limit int) ([]User, error)
	Count(ctx context.Context, search string) (int64, error)
}
