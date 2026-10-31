package dto

type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=50"`
	Description string `json:"description" binding:"max=255"`
	PolicyIDs   []int  `json:"policy_ids" binding:"required,min=1,dive,gt=0"`
}

type PermissionResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PolicyIDs   []int  `json:"policy_ids"`
}
