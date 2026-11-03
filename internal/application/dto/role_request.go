package dto

type CreateRoleRequest struct {
	Name          string `json:"name" binding:"required,min=3,max=50"`
	Description   string `json:"description" binding:"max=255"`
	IsActive      *bool  `json:"is_active,omitempty"`
	PermissionIDs []int  `json:"permission_ids" binding:"required,min=1,dive,gt=0"`
}

type RoleResponse struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	IsActive      bool   `json:"is_active"`
	IsSystem      bool   `json:"is_system"`
	PermissionIDs []int  `json:"permission_ids"`
}
