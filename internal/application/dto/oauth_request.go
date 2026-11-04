package dto

type GitHubAuthURLResponse struct {
	URL   string `json:"url"`
	State string `json:"state"`
}

type GitHubOAuthCallbackRequest struct {
	Code  string `json:"code" form:"code" binding:"required"`
	State string `json:"state" form:"state" binding:"required"`
}
