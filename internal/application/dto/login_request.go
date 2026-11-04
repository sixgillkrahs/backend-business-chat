package dto

import "github.com/sixgillkrahs/backend-business-chat/internal/domain"

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

type UserLoginResponse struct {
	ID          int                 `json:"id"`
	Username    string              `json:"username"`
	Email       string              `json:"email"`
	FullName    string              `json:"full_name"`
	AvatarURL   string              `json:"avatar_url,omitempty"`
	Role        *RoleResponse       `json:"role,omitempty"`
	Permissions []domain.Permission `json:"permissions,omitempty"`
}

type LoginResponse struct {
	AccessToken  string             `json:"access_token"`
	RefreshToken string             `json:"refresh_token,omitempty"`
	TokenType    string             `json:"token_type"`
	ExpiresIn    int64              `json:"expires_in"`
	User         *UserLoginResponse `json:"user,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}
