package dto

type PolicyRequest struct {
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	ResourceId     int    `json:"resource_id" binding:"required"`
	ActionId       int    `json:"action_id" binding:"required"`
	IdempotenceKey string `json:"idempotence_key" binding:"required"`
}

type UpdatePolicyRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	ResourceId  *int    `json:"resource_id"`
	ActionId    *int    `json:"action_id"`
}
