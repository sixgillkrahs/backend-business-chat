package dto

type PolicyRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ResourceId  int    `json:"resource_id" binding:"required"`
	ActionId    int    `json:"action_id" binding:"required"`
}
