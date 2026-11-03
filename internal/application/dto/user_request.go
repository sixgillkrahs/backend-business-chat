package dto

type CreateUserRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=50"`
	Email       string `json:"email" binding:"required,email,max=100"`
	Password    string `json:"password" binding:"required,min=8,max=72"`
	FullName    string `json:"full_name" binding:"required,min=1,max=100"`
	PhoneNumber string `json:"phone_number" binding:"omitempty,max=20"`
	Address     string `json:"address" binding:"omitempty,max=255"`
	AvatarURL   string `json:"avatar_url" binding:"omitempty,url,max=255"`
	RoleID      int    `json:"role_id" binding:"required,gt=0"`
}
