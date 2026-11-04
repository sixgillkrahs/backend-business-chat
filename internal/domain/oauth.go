package domain

import (
	"context"
	"time"
)

type OAuthAccount struct {
	ID             int       `json:"id" db:"id"`
	UserID         int       `json:"user_id" db:"user_id"`
	Provider       string    `json:"provider" db:"provider"`
	ProviderUserID string    `json:"provider_user_id" db:"provider_user_id"`
	Email          string    `json:"email" db:"email"`
	AvatarURL      string    `json:"avatar_url" db:"avatar_url"`
	AccessToken    string    `json:"-" db:"access_token"` // AES-256-GCM encrypted in database
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

func (OAuthAccount) TableName() string {
	return "oauth_accounts"
}

type OAuthRepository interface {
	FindByProviderAndUserID(ctx context.Context, provider, providerUserID string) (*OAuthAccount, error)
	Create(ctx context.Context, account *OAuthAccount) error
	UpdateAccessToken(ctx context.Context, id int, encryptedToken string) error
	FindByUserID(ctx context.Context, userID int) ([]OAuthAccount, error)
}

// GitHubRepository represents a user's repository on GitHub
type GitHubRepository struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Private       bool   `json:"private"`
	HTMLURL       string `json:"html_url"`
	Description   string `json:"description"`
	DefaultBranch string `json:"default_branch"`
}

// GitHubCreateIssueRequest represents the payload for creating an issue
type GitHubCreateIssueRequest struct {
	Title     string   `json:"title" binding:"required"`
	Body      string   `json:"body,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
}

// GitHubIssueResponse represents a created GitHub issue
type GitHubIssueResponse struct {
	ID      int64  `json:"id"`
	Number  int    `json:"number"`
	Title   string `json:"title"`
	HTMLURL string `json:"html_url"`
	State   string `json:"state"`
}
