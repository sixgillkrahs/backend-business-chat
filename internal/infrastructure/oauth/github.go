package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sixgillkrahs/backend-business-chat/internal/config"
	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
)

type GitHubUserInfo struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type GitHubEmailInfo struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

type GitHubClient interface {
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (string, error)
	GetUserInfo(ctx context.Context, accessToken string) (*GitHubUserInfo, error)
	GetUserRepositories(ctx context.Context, accessToken string) ([]*domain.GitHubRepository, error)
	CreateIssue(ctx context.Context, accessToken string, owner, repo string, req domain.GitHubCreateIssueRequest) (*domain.GitHubIssueResponse, error)
}

type gitHubClient struct {
	cfg        config.GitHubOAuthConfig
	httpClient *http.Client
}

func NewGitHubClient(cfg config.GitHubOAuthConfig) GitHubClient {
	return &gitHubClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *gitHubClient) GetAuthURL(state string) string {
	baseURL := "https://github.com/login/oauth/authorize"
	params := url.Values{}
	params.Set("client_id", c.cfg.ClientID)
	params.Set("redirect_uri", c.cfg.RedirectURI)

	scopes := strings.TrimSpace(c.cfg.Scopes)
	if scopes == "" {
		scopes = "user:email read:user repo"
	}
	params.Set("scope", scopes)
	params.Set("state", state)
	params.Set("allow_signup", "true")
	return baseURL + "?" + params.Encode()
}

func (c *gitHubClient) ExchangeCode(ctx context.Context, code string) (string, error) {
	if c.cfg.ClientID == "" || c.cfg.ClientSecret == "" {
		return "", errors.New("github oauth client_id or client_secret is not configured")
	}

	tokenURL := "https://github.com/login/oauth/access_token"
	data := url.Values{}
	data.Set("client_id", c.cfg.ClientID)
	data.Set("client_secret", c.cfg.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", c.cfg.RedirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create code exchange request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "backend-business-chat")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute code exchange request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github returned non-200 status during code exchange: %d", resp.StatusCode)
	}

	var res struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("failed to decode github token response: %w", err)
	}

	if res.Error != "" {
		return "", fmt.Errorf("%w: %s - %s", domain.ErrOAuthCodeExchangeFailed, res.Error, res.ErrorDesc)
	}

	if res.AccessToken == "" {
		return "", domain.ErrOAuthCodeExchangeFailed
	}

	return res.AccessToken, nil
}

func (c *gitHubClient) GetUserInfo(ctx context.Context, accessToken string) (*GitHubUserInfo, error) {
	userURL := "https://api.github.com/user"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "backend-business-chat")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute user info request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, domain.ErrGitHubTokenInvalid
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github returned non-200 status for user info: %d", resp.StatusCode)
	}

	var userInfo GitHubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode github user info: %w", err)
	}

	if userInfo.ID == 0 {
		return nil, domain.ErrOAuthUserNotFound
	}

	// OWASP Requirement: If email is private/empty, fetch emails and only accept verified email
	if userInfo.Email == "" {
		verifiedEmail, err := c.getPrimaryVerifiedEmail(ctx, accessToken)
		if err != nil {
			return nil, err
		}
		userInfo.Email = verifiedEmail
	}

	return &userInfo, nil
}

func (c *gitHubClient) getPrimaryVerifiedEmail(ctx context.Context, accessToken string) (string, error) {
	emailsURL := "https://api.github.com/user/emails"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, emailsURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create emails request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "backend-business-chat")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute emails request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", domain.ErrGitHubTokenInvalid
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github returned non-200 status for emails: %d", resp.StatusCode)
	}

	var emails []GitHubEmailInfo
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("failed to decode github emails: %w", err)
	}

	// Find primary and verified email
	for _, em := range emails {
		if em.Primary && em.Verified {
			return em.Email, nil
		}
	}

	// Fallback to any verified email
	for _, em := range emails {
		if em.Verified {
			return em.Email, nil
		}
	}

	return "", domain.ErrOAuthEmailNotFound
}

func (c *gitHubClient) GetUserRepositories(ctx context.Context, accessToken string) ([]*domain.GitHubRepository, error) {
	repoURL := "https://api.github.com/user/repos?sort=updated&per_page=30"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, repoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create repositories request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "backend-business-chat")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute repositories request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, domain.ErrGitHubTokenInvalid
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github returned non-200 status for repos: %d (%s)", resp.StatusCode, string(body))
	}

	var repos []*domain.GitHubRepository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode github repositories: %w", err)
	}

	return repos, nil
}

func (c *gitHubClient) CreateIssue(ctx context.Context, accessToken string, owner, repo string, issueReq domain.GitHubCreateIssueRequest) (*domain.GitHubIssueResponse, error) {
	issueURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)

	reqBytes, err := json.Marshal(issueReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal issue request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, issueURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create issue request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "backend-business-chat")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute create issue request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, domain.ErrGitHubTokenInvalid
	}

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github returned status %d creating issue: %s", resp.StatusCode, string(body))
	}

	var issueResp domain.GitHubIssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&issueResp); err != nil {
		return nil, fmt.Errorf("failed to decode github issue response: %w", err)
	}

	return &issueResp, nil
}
