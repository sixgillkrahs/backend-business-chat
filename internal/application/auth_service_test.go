package application_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sixgillkrahs/backend-business-chat/internal/application"
	"github.com/sixgillkrahs/backend-business-chat/internal/application/dto"
	"github.com/sixgillkrahs/backend-business-chat/internal/config"
	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/oauth"
	"github.com/sixgillkrahs/backend-business-chat/pkg/utils"
)

// Mock Role Repository
type mockRoleRepo struct {
	roles map[int]*domain.Role
}

func (m *mockRoleRepo) GetRoleById(ctx context.Context, id int) (*domain.Role, error) {
	if r, ok := m.roles[id]; ok {
		return r, nil
	}
	return nil, nil
}
func (m *mockRoleRepo) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	if m.roles != nil {
		for _, r := range m.roles {
			if strings.EqualFold(r.Name, name) {
				return r, nil
			}
		}
	}
	return nil, nil
}
func (m *mockRoleRepo) CreateWithPermissions(ctx context.Context, role *domain.Role, permissionIDs []int) error {
	return nil
}
func (m *mockRoleRepo) GetRolesPaging(ctx context.Context, search string, offset, limit int) ([]domain.Role, error) {
	var result []domain.Role
	for _, r := range m.roles {
		result = append(result, *r)
	}
	return result, nil
}
func (m *mockRoleRepo) Count(ctx context.Context, search string) (int64, error) { return 0, nil }
func (m *mockRoleRepo) ValidatePermissionIDs(ctx context.Context, permissionIDs []int) (bool, error) {
	return true, nil
}
func (m *mockRoleRepo) UpdateWithPermissions(ctx context.Context, role *domain.Role, permissionIDs []int) error {
	return nil
}
func (m *mockRoleRepo) DeleteRoleById(ctx context.Context, id int) error { return nil }
func (m *mockRoleRepo) CountUsersByRoleId(ctx context.Context, roleId int) (int64, error) {
	return 0, nil
}
func (m *mockRoleRepo) UpdateRoleStatus(ctx context.Context, id int, isActive bool) error {
	if r, ok := m.roles[id]; ok {
		r.IsActive = isActive
		return nil
	}
	return domain.ErrRoleNotFound
}
func (m *mockRoleRepo) DeactivateRoleById(ctx context.Context, id int) error {
	return m.UpdateRoleStatus(ctx, id, false)
}

// Mock User Repository
type mockUserRepo struct {
	users           []domain.User
	usersByID       map[int]*domain.User
	usersByUsername map[string]*domain.User
	usersByEmail    map[string]*domain.User
	createdUser     *domain.User
	createdHash     string
	updatedUser     *domain.User
	lastLimit       int
	lastOffset      int
	lastSearch      string
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error { return nil }
func (m *mockUserRepo) CreateWithAuth(ctx context.Context, user *domain.User, passwordHash string) error {
	m.createdUser = user
	m.createdHash = passwordHash
	user.ID = 100
	return nil
}
func (m *mockUserRepo) FindByID(ctx context.Context, id int) (*domain.User, error) {
	if m.usersByID != nil {
		if u, ok := m.usersByID[id]; ok {
			copied := *u
			return &copied, nil
		}
	}
	return nil, nil
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if u, ok := m.usersByEmail[email]; ok {
		return u, nil
	}
	return nil, nil
}
func (m *mockUserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	if u, ok := m.usersByUsername[username]; ok {
		return u, nil
	}
	return nil, nil
}
func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	m.updatedUser = user
	if m.usersByID != nil {
		m.usersByID[user.ID] = user
	}
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id int, deletedBy string) error {
	if m.usersByID != nil {
		if u, ok := m.usersByID[id]; ok {
			u.SoftDelete(deletedBy)
			return nil
		}
	}
	return domain.ErrUserNotFound
}
func (m *mockUserRepo) GetUsersPaging(ctx context.Context, search string, offset, limit int) ([]domain.User, error) {
	m.lastLimit = limit
	m.lastOffset = offset
	m.lastSearch = search

	var filtered []domain.User
	for _, u := range m.users {
		if search == "" || strings.Contains(strings.ToLower(u.Username), strings.ToLower(search)) || strings.Contains(strings.ToLower(u.Email), strings.ToLower(search)) {
			filtered = append(filtered, u)
		}
	}
	if offset >= len(filtered) {
		return []domain.User{}, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[offset:end], nil
}

func (m *mockUserRepo) Count(ctx context.Context, search string) (int64, error) {
	var count int64
	for _, u := range m.users {
		if search == "" || strings.Contains(strings.ToLower(u.Username), strings.ToLower(search)) || strings.Contains(strings.ToLower(u.Email), strings.ToLower(search)) {
			count++
		}
	}
	return count, nil
}

type mockAuthRepo struct {
	auths map[string]*domain.Auth
}

func (m *mockAuthRepo) FindByUsername(ctx context.Context, username string) (domain.Auth, error) {
	if m.auths != nil {
		if a, ok := m.auths[username]; ok {
			return *a, nil
		}
	}
	return domain.Auth{}, domain.ErrUserNotFound
}

type mockOAuthRepo struct {
	accounts []*domain.OAuthAccount
}

func (m *mockOAuthRepo) FindByProviderAndUserID(ctx context.Context, provider, providerUserID string) (*domain.OAuthAccount, error) {
	for _, a := range m.accounts {
		if a.Provider == provider && a.ProviderUserID == providerUserID {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockOAuthRepo) Create(ctx context.Context, account *domain.OAuthAccount) error {
	account.ID = len(m.accounts) + 1
	m.accounts = append(m.accounts, account)
	return nil
}

func (m *mockOAuthRepo) UpdateAccessToken(ctx context.Context, id int, encryptedToken string) error {
	for _, a := range m.accounts {
		if a.ID == id {
			a.AccessToken = encryptedToken
			return nil
		}
	}
	return nil
}

func (m *mockOAuthRepo) FindByUserID(ctx context.Context, userID int) ([]domain.OAuthAccount, error) {
	var result []domain.OAuthAccount
	for _, a := range m.accounts {
		if a.UserID == userID {
			result = append(result, *a)
		}
	}
	return result, nil
}

type mockGitHubClient struct {
	authURL   string
	tokenErr  error
	token     string
	userInfo  *oauth.GitHubUserInfo
	userErr   error
	repos     []*domain.GitHubRepository
	reposErr  error
	issueResp *domain.GitHubIssueResponse
	issueErr  error
}

func (m *mockGitHubClient) GetAuthURL(state string) string {
	if m.authURL != "" {
		return m.authURL + "?state=" + state
	}
	return "https://github.com/login/oauth/authorize?state=" + state
}

func (m *mockGitHubClient) ExchangeCode(ctx context.Context, code string) (string, error) {
	if m.tokenErr != nil {
		return "", m.tokenErr
	}
	if m.token != "" {
		return m.token, nil
	}
	return "mock-access-token", nil
}

func (m *mockGitHubClient) GetUserInfo(ctx context.Context, accessToken string) (*oauth.GitHubUserInfo, error) {
	if m.userErr != nil {
		return nil, m.userErr
	}
	return m.userInfo, nil
}

func (m *mockGitHubClient) GetUserRepositories(ctx context.Context, accessToken string) ([]*domain.GitHubRepository, error) {
	if m.reposErr != nil {
		return nil, m.reposErr
	}
	return m.repos, nil
}

func (m *mockGitHubClient) CreateIssue(ctx context.Context, accessToken string, owner, repo string, req domain.GitHubCreateIssueRequest) (*domain.GitHubIssueResponse, error) {
	if m.issueErr != nil {
		return nil, m.issueErr
	}
	return m.issueResp, nil
}

func setupService(roleRepo domain.RoleRepository, userRepo domain.UserRepository) *application.AuthService {
	cfg := &config.Config{}
	return application.NewAuthService(cfg, nil, nil, nil, nil, nil, roleRepo, userRepo)
}

func setupFullService(authRepo domain.AuthRepository, roleRepo domain.RoleRepository, userRepo domain.UserRepository) *application.AuthService {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Jwt: config.JWT{
				SecretKey: "supersecretkeyforunittestingjwttoken",
				Duration:  60,
			},
		},
	}
	return application.NewAuthService(cfg, nil, authRepo, nil, nil, nil, roleRepo, userRepo)
}

func TestCreateUser_Success(t *testing.T) {
	roleRepo := &mockRoleRepo{
		roles: map[int]*domain.Role{
			1: {ID: 1, Name: "Admin", IsActive: true},
		},
	}
	userRepo := &mockUserRepo{
		usersByUsername: make(map[string]*domain.User),
		usersByEmail:    make(map[string]*domain.User),
	}
	service := setupService(roleRepo, userRepo)

	req := dto.CreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "SecurePassword123!",
		FullName: "New User",
		RoleID:   1,
	}

	user, err := service.CreateUser(context.Background(), req, "vi", "superadmin")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if user.ID != 100 {
		t.Errorf("expected user ID 100, got: %d", user.ID)
	}
	if user.Username != "newuser" {
		t.Errorf("expected username newuser, got: %s", user.Username)
	}
	if user.Email != "newuser@example.com" {
		t.Errorf("expected email newuser@example.com, got: %s", user.Email)
	}
	if user.Role == nil || user.Role.ID != 1 {
		t.Errorf("expected role to be populated")
	}

	if !utils.CheckPasswordHash("SecurePassword123!", userRepo.createdHash) {
		t.Errorf("password hash does not match original password via bcrypt")
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	roleRepo := &mockRoleRepo{
		roles: map[int]*domain.Role{1: {ID: 1, Name: "Admin"}},
	}
	userRepo := &mockUserRepo{
		usersByUsername: map[string]*domain.User{
			"existing": {ID: 1, Username: "existing"},
		},
		usersByEmail: make(map[string]*domain.User),
	}
	service := setupService(roleRepo, userRepo)

	req := dto.CreateUserRequest{
		Username: "existing",
		Email:    "other@example.com",
		Password: "SecurePassword123!",
		FullName: "Other User",
		RoleID:   1,
	}

	_, err := service.CreateUser(context.Background(), req, "vi", "admin")
	if err != domain.ErrUsernameDuplicate {
		t.Fatalf("expected ErrUsernameDuplicate, got: %v", err)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	roleRepo := &mockRoleRepo{
		roles: map[int]*domain.Role{1: {ID: 1, Name: "Admin"}},
	}
	userRepo := &mockUserRepo{
		usersByUsername: make(map[string]*domain.User),
		usersByEmail: map[string]*domain.User{
			"existing@example.com": {ID: 1, Email: "existing@example.com"},
		},
	}
	service := setupService(roleRepo, userRepo)

	req := dto.CreateUserRequest{
		Username: "otheruser",
		Email:    "existing@example.com",
		Password: "SecurePassword123!",
		FullName: "Other User",
		RoleID:   1,
	}

	_, err := service.CreateUser(context.Background(), req, "vi", "admin")
	if err != domain.ErrEmailDuplicate {
		t.Fatalf("expected ErrEmailDuplicate, got: %v", err)
	}
}

func TestCreateUser_RoleNotFound(t *testing.T) {
	roleRepo := &mockRoleRepo{roles: map[int]*domain.Role{}}
	userRepo := &mockUserRepo{
		usersByUsername: make(map[string]*domain.User),
		usersByEmail:    make(map[string]*domain.User),
	}
	service := setupService(roleRepo, userRepo)

	req := dto.CreateUserRequest{
		Username: "validuser",
		Email:    "valid@example.com",
		Password: "SecurePassword123!",
		FullName: "Valid User",
		RoleID:   999,
	}

	_, err := service.CreateUser(context.Background(), req, "vi", "admin")
	if err != domain.ErrRoleNotFound {
		t.Fatalf("expected ErrRoleNotFound, got: %v", err)
	}
}

func TestCreateUser_PasswordTooShort(t *testing.T) {
	roleRepo := &mockRoleRepo{roles: map[int]*domain.Role{1: {ID: 1}}}
	userRepo := &mockUserRepo{}
	service := setupService(roleRepo, userRepo)

	req := dto.CreateUserRequest{
		Username: "validuser",
		Email:    "valid@example.com",
		Password: "123",
		FullName: "Valid User",
		RoleID:   1,
	}

	_, err := service.CreateUser(context.Background(), req, "vi", "admin")
	if err != domain.ErrPasswordTooShort {
		t.Fatalf("expected ErrPasswordTooShort, got: %v", err)
	}
}

func TestGetUsersPaging_DefaultPagination(t *testing.T) {
	users := make([]domain.User, 25)
	for i := 0; i < 25; i++ {
		users[i] = domain.User{
			ID:       i + 1,
			Username: fmt.Sprintf("user_%d", i+1),
			Email:    fmt.Sprintf("user_%d@example.com", i+1),
			FullName: fmt.Sprintf("User %d", i+1),
		}
	}
	userRepo := &mockUserRepo{users: users}
	service := setupService(&mockRoleRepo{}, userRepo)

	resp, err := service.GetUsersPaging(context.Background(), 0, 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Total != 25 {
		t.Errorf("expected total 25, got: %d", resp.Total)
	}
	if resp.Limit != 10 {
		t.Errorf("expected limit 10, got: %d", resp.Limit)
	}
	if resp.Offset != 0 {
		t.Errorf("expected offset 0, got: %d", resp.Offset)
	}
	if len(resp.Items) != 10 {
		t.Errorf("expected 10 items, got: %d", len(resp.Items))
	}
}

func TestGetUsersPaging_CustomLimitAndOffset(t *testing.T) {
	users := make([]domain.User, 25)
	for i := 0; i < 25; i++ {
		users[i] = domain.User{
			ID:       i + 1,
			Username: fmt.Sprintf("user_%d", i+1),
			Email:    fmt.Sprintf("user_%d@example.com", i+1),
		}
	}
	userRepo := &mockUserRepo{users: users}
	service := setupService(&mockRoleRepo{}, userRepo)

	resp, err := service.GetUsersPaging(context.Background(), 5, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Limit != 5 {
		t.Errorf("expected limit 5, got: %d", resp.Limit)
	}
	if resp.Offset != 10 {
		t.Errorf("expected offset 10, got: %d", resp.Offset)
	}
	if len(resp.Items) != 5 {
		t.Errorf("expected 5 items, got: %d", len(resp.Items))
	}
	if resp.Items[0].ID != 11 {
		t.Errorf("expected first item ID 11, got: %d", resp.Items[0].ID)
	}
}

func TestGetUsersPaging_BoundsValidation(t *testing.T) {
	userRepo := &mockUserRepo{users: []domain.User{}}
	service := setupService(&mockRoleRepo{}, userRepo)

	resp, err := service.GetUsersPaging(context.Background(), 100, -5, strings.Repeat("a", 150))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Limit != 50 {
		t.Errorf("expected limit clamped to 50, got: %d", resp.Limit)
	}
	if resp.Offset != 0 {
		t.Errorf("expected negative offset clamped to 0, got: %d", resp.Offset)
	}
	if len(userRepo.lastSearch) != 100 {
		t.Errorf("expected search truncated to 100 chars, got length: %d", len(userRepo.lastSearch))
	}
}

func TestGetUsersPaging_Search(t *testing.T) {
	users := []domain.User{
		{ID: 1, Username: "alice", Email: "alice@company.com"},
		{ID: 2, Username: "bob", Email: "bob@company.com"},
		{ID: 3, Username: "charlie", Email: "charlie@other.com"},
	}
	userRepo := &mockUserRepo{users: users}
	service := setupService(&mockRoleRepo{}, userRepo)

	resp, err := service.GetUsersPaging(context.Background(), 10, 0, "company")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Total != 2 {
		t.Errorf("expected total 2, got: %d", resp.Total)
	}
	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items, got: %d", len(resp.Items))
	}
}

func TestGetUserById(t *testing.T) {
	initialUser := &domain.User{
		ID:       10,
		Username: "existing_user",
		Email:    "existing@example.com",
		FullName: "Existing User",
		IsActive: true,
	}
	userRepo := &mockUserRepo{
		usersByID: map[int]*domain.User{10: initialUser},
	}
	service := setupService(&mockRoleRepo{}, userRepo)

	u, err := service.GetUserById(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if u.ID != 10 || u.Username != "existing_user" {
		t.Errorf("expected user ID 10, got %d", u.ID)
	}

	_, err = service.GetUserById(context.Background(), 999)
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}

	_, err = service.GetUserById(context.Background(), 0)
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUpdateUserById_Success(t *testing.T) {
	currentRoleID := 1
	initialUser := &domain.User{
		ID:       10,
		Username: "oldusername",
		Email:    "old@example.com",
		FullName: "Old Name",
		RoleID:   &currentRoleID,
		IsActive: true,
	}

	roleRepo := &mockRoleRepo{
		roles: map[int]*domain.Role{
			1: {ID: 1, Name: "Admin", IsActive: true},
			2: {ID: 2, Name: "Manager", IsActive: true},
		},
	}
	userRepo := &mockUserRepo{
		usersByID:       map[int]*domain.User{10: initialUser},
		usersByUsername: map[string]*domain.User{"oldusername": initialUser},
		usersByEmail:    map[string]*domain.User{"old@example.com": initialUser},
	}
	service := setupService(roleRepo, userRepo)

	newRoleID := 2
	newFullName := "Updated Name"
	newEmail := "updated@example.com"
	newPhone := "0912345678"
	newAddress := "123 Main St"
	newActive := false

	req := dto.UpdateUserRequest{
		FullName:    &newFullName,
		Email:       &newEmail,
		PhoneNumber: &newPhone,
		Address:     &newAddress,
		RoleID:      &newRoleID,
		IsActive:    &newActive,
	}

	updated, err := service.UpdateUserById(context.Background(), 10, req, "vi", "supervisor")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if updated.FullName != "Updated Name" {
		t.Errorf("expected FullName 'Updated Name', got %s", updated.FullName)
	}
	if updated.Email != "updated@example.com" {
		t.Errorf("expected Email 'updated@example.com', got %s", updated.Email)
	}
	if updated.RoleID == nil || *updated.RoleID != 2 {
		t.Errorf("expected RoleID 2, got %v", updated.RoleID)
	}
	if updated.Role == nil || updated.Role.Name != "Manager" {
		t.Errorf("expected Role populated with Manager, got %v", updated.Role)
	}
	if updated.IsActive != false {
		t.Errorf("expected IsActive false")
	}
	if updated.UpdatedBy != "supervisor" {
		t.Errorf("expected UpdatedBy supervisor, got %s", updated.UpdatedBy)
	}
}

func TestUpdateUserById_UserNotFound(t *testing.T) {
	service := setupService(&mockRoleRepo{}, &mockUserRepo{usersByID: map[int]*domain.User{}})

	newName := "Name"
	req := dto.UpdateUserRequest{FullName: &newName}

	_, err := service.UpdateUserById(context.Background(), 999, req, "vi", "admin")
	if err != domain.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUpdateUserById_DuplicateUsername(t *testing.T) {
	user1 := &domain.User{ID: 1, Username: "alice", Email: "alice@example.com", FullName: "Alice"}
	user2 := &domain.User{ID: 2, Username: "bob", Email: "bob@example.com", FullName: "Bob"}

	userRepo := &mockUserRepo{
		usersByID:       map[int]*domain.User{1: user1, 2: user2},
		usersByUsername: map[string]*domain.User{"alice": user1, "bob": user2},
		usersByEmail:    map[string]*domain.User{"alice@example.com": user1, "bob@example.com": user2},
	}
	service := setupService(&mockRoleRepo{}, userRepo)

	targetUsername := "bob"
	req := dto.UpdateUserRequest{Username: &targetUsername}

	_, err := service.UpdateUserById(context.Background(), 1, req, "vi", "admin")
	if err != domain.ErrUsernameDuplicate {
		t.Fatalf("expected ErrUsernameDuplicate, got %v", err)
	}
}

func TestUpdateUserById_DuplicateEmail(t *testing.T) {
	user1 := &domain.User{ID: 1, Username: "alice", Email: "alice@example.com", FullName: "Alice"}
	user2 := &domain.User{ID: 2, Username: "bob", Email: "bob@example.com", FullName: "Bob"}

	userRepo := &mockUserRepo{
		usersByID:       map[int]*domain.User{1: user1, 2: user2},
		usersByUsername: map[string]*domain.User{"alice": user1, "bob": user2},
		usersByEmail:    map[string]*domain.User{"alice@example.com": user1, "bob@example.com": user2},
	}
	service := setupService(&mockRoleRepo{}, userRepo)

	targetEmail := "bob@example.com"
	req := dto.UpdateUserRequest{Email: &targetEmail}

	_, err := service.UpdateUserById(context.Background(), 1, req, "vi", "admin")
	if err != domain.ErrEmailDuplicate {
		t.Fatalf("expected ErrEmailDuplicate, got %v", err)
	}
}

func TestUpdateUserById_RoleNotFound(t *testing.T) {
	user1 := &domain.User{ID: 1, Username: "alice", Email: "alice@example.com", FullName: "Alice"}

	roleRepo := &mockRoleRepo{roles: map[int]*domain.Role{}}
	userRepo := &mockUserRepo{
		usersByID: map[int]*domain.User{1: user1},
	}
	service := setupService(roleRepo, userRepo)

	invalidRoleID := 999
	req := dto.UpdateUserRequest{RoleID: &invalidRoleID}

	_, err := service.UpdateUserById(context.Background(), 1, req, "vi", "admin")
	if err != domain.ErrRoleNotFound {
		t.Fatalf("expected ErrRoleNotFound, got %v", err)
	}
}

func TestDeactivateRoleById_Success(t *testing.T) {
	role := &domain.Role{ID: 10, Name: "Tester", IsActive: true, IsSystem: false}
	roleRepo := &mockRoleRepo{roles: map[int]*domain.Role{10: role}}
	service := setupService(roleRepo, &mockUserRepo{})

	err := service.DeactivateRoleById(context.Background(), 10, "vi")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if role.IsActive != false {
		t.Errorf("expected role.IsActive to be false, got true")
	}
}

func TestDeactivateRoleById_SystemRole(t *testing.T) {
	systemRole := &domain.Role{ID: 1, Name: "Admin", IsActive: true, IsSystem: true}
	roleRepo := &mockRoleRepo{roles: map[int]*domain.Role{1: systemRole}}
	service := setupService(roleRepo, &mockUserRepo{})

	err := service.DeactivateRoleById(context.Background(), 1, "vi")
	if err != domain.ErrSystemRoleCannotDeactivate {
		t.Fatalf("expected ErrSystemRoleCannotDeactivate, got %v", err)
	}

	if systemRole.IsActive != true {
		t.Errorf("expected systemRole.IsActive to remain true")
	}
}

func TestDeactivateRoleById_NotFound(t *testing.T) {
	roleRepo := &mockRoleRepo{roles: map[int]*domain.Role{}}
	service := setupService(roleRepo, &mockUserRepo{})

	err := service.DeactivateRoleById(context.Background(), 999, "vi")
	if err != domain.ErrRoleNotFound {
		t.Fatalf("expected ErrRoleNotFound, got %v", err)
	}
}

func TestDeactivateRoleById_InvalidID(t *testing.T) {
	service := setupService(&mockRoleRepo{}, &mockUserRepo{})

	err := service.DeactivateRoleById(context.Background(), 0, "vi")
	if err != domain.ErrRoleNotFound {
		t.Fatalf("expected ErrRoleNotFound for id <= 0, got %v", err)
	}
}

func TestUpdateRoleStatus_Toggle(t *testing.T) {
	role := &domain.Role{ID: 20, Name: "Moderator", IsActive: true, IsSystem: false}
	roleRepo := &mockRoleRepo{roles: map[int]*domain.Role{20: role}}
	service := setupService(roleRepo, &mockUserRepo{})

	// 1st toggle: true -> false (no is_active specified)
	updated, err := service.UpdateRoleStatus(context.Background(), 20, dto.UpdateRoleStatusRequest{}, "vi")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if updated.IsActive != false {
		t.Errorf("expected role.IsActive to be false after toggle, got %v", updated.IsActive)
	}

	// 2nd toggle: false -> true
	updated, err = service.UpdateRoleStatus(context.Background(), 20, dto.UpdateRoleStatusRequest{}, "vi")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if updated.IsActive != true {
		t.Errorf("expected role.IsActive to be true after toggle, got %v", updated.IsActive)
	}
}

func TestUpdateRoleStatus_ExplicitValue(t *testing.T) {
	role := &domain.Role{ID: 21, Name: "Editor", IsActive: true, IsSystem: false}
	roleRepo := &mockRoleRepo{roles: map[int]*domain.Role{21: role}}
	service := setupService(roleRepo, &mockUserRepo{})

	isActiveFalse := false
	updated, err := service.UpdateRoleStatus(context.Background(), 21, dto.UpdateRoleStatusRequest{IsActive: &isActiveFalse}, "vi")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if updated.IsActive != false {
		t.Errorf("expected role.IsActive to be false, got %v", updated.IsActive)
	}

	isActiveTrue := true
	updated, err = service.UpdateRoleStatus(context.Background(), 21, dto.UpdateRoleStatusRequest{IsActive: &isActiveTrue}, "vi")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if updated.IsActive != true {
		t.Errorf("expected role.IsActive to be true, got %v", updated.IsActive)
	}
}

func TestUpdateRoleStatus_SystemRole(t *testing.T) {
	systemRole := &domain.Role{ID: 1, Name: "Admin", IsActive: true, IsSystem: true}
	roleRepo := &mockRoleRepo{roles: map[int]*domain.Role{1: systemRole}}
	service := setupService(roleRepo, &mockUserRepo{})

	_, err := service.UpdateRoleStatus(context.Background(), 1, dto.UpdateRoleStatusRequest{}, "vi")
	if err != domain.ErrSystemRoleProtected {
		t.Fatalf("expected ErrSystemRoleProtected for system role, got %v", err)
	}
}

func TestDeleteUserById_Success(t *testing.T) {
	user := &domain.User{ID: 5, Username: "regular_user", Email: "regular@example.com"}
	userRepo := &mockUserRepo{usersByID: map[int]*domain.User{5: user}}
	service := setupService(&mockRoleRepo{}, userRepo)

	err := service.DeleteUserById(context.Background(), 5, "vi", "admin_actor")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !user.IsDeleted {
		t.Errorf("expected user.IsDeleted to be true")
	}
	if user.DeletedBy == nil || *user.DeletedBy != "admin_actor" {
		t.Errorf("expected DeletedBy to be 'admin_actor', got %v", user.DeletedBy)
	}
}

func TestDeleteUserById_NotFound(t *testing.T) {
	userRepo := &mockUserRepo{usersByID: map[int]*domain.User{}}
	service := setupService(&mockRoleRepo{}, userRepo)

	err := service.DeleteUserById(context.Background(), 999, "vi", "admin_actor")
	if err != domain.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestDeleteUserById_SystemUser_Admin(t *testing.T) {
	adminUser := &domain.User{ID: 1, Username: "admin", Email: "admin@example.com"}
	userRepo := &mockUserRepo{usersByID: map[int]*domain.User{1: adminUser}}
	service := setupService(&mockRoleRepo{}, userRepo)

	err := service.DeleteUserById(context.Background(), 1, "vi", "other_admin")
	if err != domain.ErrSystemUserCannotDelete {
		t.Fatalf("expected ErrSystemUserCannotDelete, got %v", err)
	}
}

func TestDeleteUserById_SystemUser_Role(t *testing.T) {
	systemUser := &domain.User{
		ID:       2,
		Username: "supermod",
		Role:     &domain.Role{ID: 1, Name: "SuperAdmin", IsSystem: true},
	}
	userRepo := &mockUserRepo{usersByID: map[int]*domain.User{2: systemUser}}
	service := setupService(&mockRoleRepo{}, userRepo)

	err := service.DeleteUserById(context.Background(), 2, "vi", "other_admin")
	if err != domain.ErrSystemUserCannotDelete {
		t.Fatalf("expected ErrSystemUserCannotDelete, got %v", err)
	}
}

func TestDeleteUserById_CannotDeleteSelf(t *testing.T) {
	user := &domain.User{ID: 10, Username: "operator", Email: "operator@example.com"}
	userRepo := &mockUserRepo{usersByID: map[int]*domain.User{10: user}}
	service := setupService(&mockRoleRepo{}, userRepo)

	err := service.DeleteUserById(context.Background(), 10, "vi", "operator")
	if err != domain.ErrCannotDeleteSelf {
		t.Fatalf("expected ErrCannotDeleteSelf, got %v", err)
	}
}

func TestDeleteUserById_InvalidID(t *testing.T) {
	service := setupService(&mockRoleRepo{}, &mockUserRepo{})

	err := service.DeleteUserById(context.Background(), 0, "vi", "admin")
	if err != domain.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound for id <= 0, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	hashedPassword, _ := utils.HashPassword("123456aA@1")
	auth := &domain.Auth{
		UserId:       1,
		Username:     "admin",
		PasswordHash: hashedPassword,
		RoleId:       1,
		IsActive:     true,
		IsDeleted:    false,
	}
	user := &domain.User{
		ID:        1,
		Username:  "admin",
		Email:     "admin@example.com",
		FullName:  "System Administrator",
		IsActive:  true,
		IsDeleted: false,
	}
	role := &domain.Role{
		ID:       1,
		Name:     "Admin",
		IsActive: true,
		IsSystem: true,
		Permissions: []domain.Permission{
			{ID: 1, Name: "USER_READ"},
		},
	}

	authRepo := &mockAuthRepo{auths: map[string]*domain.Auth{"admin": auth}}
	userRepo := &mockUserRepo{usersByID: map[int]*domain.User{1: user}}
	roleRepo := &mockRoleRepo{roles: map[int]*domain.Role{1: role}}
	service := setupFullService(authRepo, roleRepo, userRepo)

	resp, err := service.Login(context.Background(), dto.LoginRequest{
		Username: "admin",
		Password: "123456aA@1",
	}, "vi", "127.0.0.1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp == nil || resp.AccessToken == "" {
		t.Fatalf("expected non-empty access token")
	}
	if resp.RefreshToken == "" {
		t.Fatalf("expected non-empty refresh token")
	}
	if resp.TokenType != "Bearer" {
		t.Errorf("expected TokenType 'Bearer', got %s", resp.TokenType)
	}
	if resp.ExpiresIn <= 0 {
		t.Errorf("expected positive ExpiresIn, got %d", resp.ExpiresIn)
	}
	if resp.User == nil || resp.User.Username != "admin" {
		t.Errorf("expected User.Username 'admin', got %v", resp.User)
	}
	if resp.User.Role == nil || resp.User.Role.Name != "Admin" {
		t.Errorf("expected User.Role.Name 'Admin'")
	}
	if len(resp.User.Permissions) != 1 {
		t.Errorf("expected 1 permission, got %d", len(resp.User.Permissions))
	}
}

func TestLogin_InvalidUsername(t *testing.T) {
	authRepo := &mockAuthRepo{auths: map[string]*domain.Auth{}}
	service := setupFullService(authRepo, &mockRoleRepo{}, &mockUserRepo{})

	_, err := service.Login(context.Background(), dto.LoginRequest{
		Username: "nonexistent",
		Password: "Password123@",
	}, "vi", "127.0.0.1")
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials for non-existent user, got %v", err)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	hashedPassword, _ := utils.HashPassword("CorrectPassword123@")
	auth := &domain.Auth{
		UserId:       1,
		Username:     "testuser",
		PasswordHash: hashedPassword,
		RoleId:       1,
		IsActive:     true,
	}

	authRepo := &mockAuthRepo{auths: map[string]*domain.Auth{"testuser": auth}}
	service := setupFullService(authRepo, &mockRoleRepo{}, &mockUserRepo{})

	_, err := service.Login(context.Background(), dto.LoginRequest{
		Username: "testuser",
		Password: "WrongPassword123@",
	}, "vi", "127.0.0.1")
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials for wrong password, got %v", err)
	}
}

func TestLogin_UserLocked(t *testing.T) {
	hashedPassword, _ := utils.HashPassword("Password123@")
	auth := &domain.Auth{
		UserId:       1,
		Username:     "lockeduser",
		PasswordHash: hashedPassword,
		RoleId:       1,
		IsActive:     false, // Locked
	}

	authRepo := &mockAuthRepo{auths: map[string]*domain.Auth{"lockeduser": auth}}
	service := setupFullService(authRepo, &mockRoleRepo{}, &mockUserRepo{})

	_, err := service.Login(context.Background(), dto.LoginRequest{
		Username: "lockeduser",
		Password: "Password123@",
	}, "vi", "127.0.0.1")
	if err != domain.ErrUserLocked {
		t.Fatalf("expected ErrUserLocked for inactive auth, got %v", err)
	}
}

func TestLogin_RoleInactive(t *testing.T) {
	hashedPassword, _ := utils.HashPassword("Password123@")
	auth := &domain.Auth{
		UserId:       1,
		Username:     "roleuser",
		PasswordHash: hashedPassword,
		RoleId:       2,
		IsActive:     true,
	}
	user := &domain.User{ID: 1, Username: "roleuser", IsActive: true}
	role := &domain.Role{ID: 2, Name: "DisabledRole", IsActive: false}

	authRepo := &mockAuthRepo{auths: map[string]*domain.Auth{"roleuser": auth}}
	userRepo := &mockUserRepo{usersByID: map[int]*domain.User{1: user}}
	roleRepo := &mockRoleRepo{roles: map[int]*domain.Role{2: role}}
	service := setupFullService(authRepo, roleRepo, userRepo)

	_, err := service.Login(context.Background(), dto.LoginRequest{
		Username: "roleuser",
		Password: "Password123@",
	}, "vi", "127.0.0.1")
	if err != domain.ErrRoleLocked {
		t.Fatalf("expected ErrRoleLocked for inactive role, got %v", err)
	}
}

func TestRefreshToken_Success(t *testing.T) {
	auth := &domain.Auth{
		UserId:    1,
		Username:  "admin",
		RoleId:    1,
		IsActive:  true,
		IsDeleted: false,
	}
	user := &domain.User{
		ID:        1,
		Username:  "admin",
		Email:     "admin@example.com",
		FullName:  "System Administrator",
		IsActive:  true,
		IsDeleted: false,
	}
	role := &domain.Role{
		ID:       1,
		Name:     "Admin",
		IsActive: true,
		Permissions: []domain.Permission{
			{ID: 1, Name: "USER_READ"},
		},
	}

	authRepo := &mockAuthRepo{auths: map[string]*domain.Auth{"admin": auth}}
	userRepo := &mockUserRepo{usersByID: map[int]*domain.User{1: user}}
	roleRepo := &mockRoleRepo{roles: map[int]*domain.Role{1: role}}
	service := setupFullService(authRepo, roleRepo, userRepo)

	// Generate a valid refresh token using utils.GenerateRefreshToken
	refreshToken, err := utils.GenerateRefreshToken(1, "admin", "Admin", "127.0.0.1", "supersecretkeyforunittestingjwttoken", 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	resp, err := service.RefreshToken(context.Background(), refreshToken, "vi", "127.0.0.1")
	if err != nil {
		t.Fatalf("expected nil error on refresh, got: %v", err)
	}
	if resp == nil || resp.AccessToken == "" {
		t.Fatalf("expected non-empty access token")
	}
	if resp.RefreshToken == "" {
		t.Fatalf("expected non-empty rotated refresh token")
	}
	if resp.User == nil || resp.User.Username != "admin" {
		t.Errorf("expected username admin, got %v", resp.User)
	}
	if len(resp.User.Permissions) != 1 {
		t.Errorf("expected 1 permission, got %d", len(resp.User.Permissions))
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	service := setupFullService(&mockAuthRepo{}, &mockRoleRepo{}, &mockUserRepo{})

	_, err := service.RefreshToken(context.Background(), "invalid.token.string", "vi", "127.0.0.1")
	if err != domain.ErrInvalidRefreshToken {
		t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestRefreshToken_AccessTokenRejectedAsRefreshToken(t *testing.T) {
	authRepo := &mockAuthRepo{auths: map[string]*domain.Auth{
		"admin": {UserId: 1, Username: "admin", RoleId: 1, IsActive: true},
	}}
	service := setupFullService(authRepo, &mockRoleRepo{}, &mockUserRepo{})

	// Generate an ACCESS token
	accessToken, err := utils.GenerateAccessToken(1, "admin", "Admin", "127.0.0.1", "supersecretkeyforunittestingjwttoken", 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	// Calling RefreshToken with an Access Token must fail with ErrInvalidRefreshToken
	_, err = service.RefreshToken(context.Background(), accessToken, "vi", "127.0.0.1")
	if err != domain.ErrInvalidRefreshToken {
		t.Fatalf("expected ErrInvalidRefreshToken when access token is provided as refresh token, got %v", err)
	}
}

func TestRefreshToken_UserLocked(t *testing.T) {
	auth := &domain.Auth{
		UserId:    1,
		Username:  "lockeduser",
		RoleId:    1,
		IsActive:  false, // Locked
		IsDeleted: false,
	}
	authRepo := &mockAuthRepo{auths: map[string]*domain.Auth{"lockeduser": auth}}
	service := setupFullService(authRepo, &mockRoleRepo{}, &mockUserRepo{})

	refreshToken, err := utils.GenerateRefreshToken(1, "lockeduser", "User", "127.0.0.1", "supersecretkeyforunittestingjwttoken", 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	_, err = service.RefreshToken(context.Background(), refreshToken, "vi", "127.0.0.1")
	if err != domain.ErrUserLocked {
		t.Fatalf("expected ErrUserLocked, got %v", err)
	}
}

func TestRefreshToken_RoleInactive(t *testing.T) {
	auth := &domain.Auth{
		UserId:    1,
		Username:  "roleuser",
		RoleId:    2,
		IsActive:  true,
		IsDeleted: false,
	}
	user := &domain.User{ID: 1, Username: "roleuser", IsActive: true}
	role := &domain.Role{ID: 2, Name: "DisabledRole", IsActive: false}

	authRepo := &mockAuthRepo{auths: map[string]*domain.Auth{"roleuser": auth}}
	userRepo := &mockUserRepo{usersByID: map[int]*domain.User{1: user}}
	roleRepo := &mockRoleRepo{roles: map[int]*domain.Role{2: role}}
	service := setupFullService(authRepo, roleRepo, userRepo)

	refreshToken, err := utils.GenerateRefreshToken(1, "roleuser", "DisabledRole", "127.0.0.1", "supersecretkeyforunittestingjwttoken", 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	_, err = service.RefreshToken(context.Background(), refreshToken, "vi", "127.0.0.1")
	if err != domain.ErrRoleLocked {
		t.Fatalf("expected ErrRoleLocked, got %v", err)
	}
}

func TestRegister_Success(t *testing.T) {
	roleRepo := &mockRoleRepo{
		roles: map[int]*domain.Role{
			2: {ID: 2, Name: "User", IsActive: true},
		},
	}
	userRepo := &mockUserRepo{
		usersByUsername: make(map[string]*domain.User),
		usersByEmail:    make(map[string]*domain.User),
	}
	service := setupService(roleRepo, userRepo)

	req := dto.RegisterRequest{
		Username: "newregistereduser",
		Email:    "registered@example.com",
		Password: "SecurePassword123@",
		FullName: "Registered User",
	}

	user, err := service.Register(context.Background(), req, "vi")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if user == nil || user.ID != 100 {
		t.Errorf("expected user ID 100, got: %v", user)
	}
	if user.Username != "newregistereduser" {
		t.Errorf("expected username newregistereduser, got: %s", user.Username)
	}
	if user.Email != "registered@example.com" {
		t.Errorf("expected email registered@example.com, got: %s", user.Email)
	}
	if user.Role == nil || user.Role.ID != 2 {
		t.Errorf("expected role ID 2, got: %v", user.Role)
	}
	if user.RoleID == nil || *user.RoleID != 2 {
		t.Errorf("expected RoleID 2, got: %v", user.RoleID)
	}
	if !utils.CheckPasswordHash("SecurePassword123@", userRepo.createdHash) {
		t.Errorf("password hash does not match original password via bcrypt")
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	roleRepo := &mockRoleRepo{
		roles: map[int]*domain.Role{2: {ID: 2, Name: "User", IsActive: true}},
	}
	userRepo := &mockUserRepo{}
	service := setupService(roleRepo, userRepo)

	req := dto.RegisterRequest{
		Username: "validuser",
		Email:    "valid@example.com",
		Password: "alllowercasepassword",
		FullName: "Valid Name",
	}

	_, err := service.Register(context.Background(), req, "vi")
	if err != domain.ErrPasswordTooWeak {
		t.Fatalf("expected ErrPasswordTooWeak, got: %v", err)
	}
}

func TestRegister_PasswordTooShort(t *testing.T) {
	roleRepo := &mockRoleRepo{
		roles: map[int]*domain.Role{2: {ID: 2, Name: "User", IsActive: true}},
	}
	userRepo := &mockUserRepo{}
	service := setupService(roleRepo, userRepo)

	req := dto.RegisterRequest{
		Username: "validuser",
		Email:    "valid@example.com",
		Password: "123",
		FullName: "Valid Name",
	}

	_, err := service.Register(context.Background(), req, "vi")
	if err != domain.ErrPasswordTooShort {
		t.Fatalf("expected ErrPasswordTooShort, got: %v", err)
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	roleRepo := &mockRoleRepo{
		roles: map[int]*domain.Role{2: {ID: 2, Name: "User", IsActive: true}},
	}
	userRepo := &mockUserRepo{
		usersByUsername: map[string]*domain.User{
			"existinguser": {ID: 1, Username: "existinguser"},
		},
		usersByEmail: make(map[string]*domain.User),
	}
	service := setupService(roleRepo, userRepo)

	req := dto.RegisterRequest{
		Username: "existinguser",
		Email:    "new@example.com",
		Password: "SecurePassword123@",
		FullName: "New User",
	}

	_, err := service.Register(context.Background(), req, "vi")
	if err != domain.ErrUsernameDuplicate {
		t.Fatalf("expected ErrUsernameDuplicate, got: %v", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	roleRepo := &mockRoleRepo{
		roles: map[int]*domain.Role{2: {ID: 2, Name: "User", IsActive: true}},
	}
	userRepo := &mockUserRepo{
		usersByUsername: make(map[string]*domain.User),
		usersByEmail: map[string]*domain.User{
			"existing@example.com": {ID: 1, Email: "existing@example.com"},
		},
	}
	service := setupService(roleRepo, userRepo)

	req := dto.RegisterRequest{
		Username: "brandnewuser",
		Email:    "existing@example.com",
		Password: "SecurePassword123@",
		FullName: "New User",
	}

	_, err := service.Register(context.Background(), req, "vi")
	if err != domain.ErrEmailDuplicate {
		t.Fatalf("expected ErrEmailDuplicate, got: %v", err)
	}
}

func TestRegister_NoDefaultRole(t *testing.T) {
	roleRepo := &mockRoleRepo{roles: map[int]*domain.Role{}}
	userRepo := &mockUserRepo{
		usersByUsername: make(map[string]*domain.User),
		usersByEmail:    make(map[string]*domain.User),
	}
	service := setupService(roleRepo, userRepo)

	req := dto.RegisterRequest{
		Username: "brandnewuser",
		Email:    "brandnew@example.com",
		Password: "SecurePassword123@",
		FullName: "New User",
	}

	_, err := service.Register(context.Background(), req, "vi")
	if err != domain.ErrRoleNotFound {
		t.Fatalf("expected ErrRoleNotFound when no active roles exist, got: %v", err)
	}
}

func TestGetGitHubAuthURL(t *testing.T) {
	service := setupFullService(&mockAuthRepo{}, &mockRoleRepo{}, &mockUserRepo{})
	ghMock := &mockGitHubClient{authURL: "https://github.com/login/oauth/authorize"}
	service.SetGitHubClient(ghMock)

	url, state, err := service.GetGitHubAuthURL(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if url == "" || state == "" {
		t.Fatalf("expected non-empty url and state")
	}
	if !strings.Contains(url, "https://github.com/login/oauth/authorize") {
		t.Errorf("expected url to contain authorize endpoint")
	}
	if !utils.ValidateOAuthState(state, "supersecretkeyforunittestingjwttoken", 5*time.Minute) {
		t.Errorf("expected valid hmac signed state")
	}
}

func TestLoginOrRegisterWithGitHub_InvalidState(t *testing.T) {
	service := setupFullService(&mockAuthRepo{}, &mockRoleRepo{}, &mockUserRepo{})
	service.SetGitHubClient(&mockGitHubClient{})
	service.SetOAuthRepo(&mockOAuthRepo{})

	_, err := service.LoginOrRegisterWithGitHub(context.Background(), "some-code", "invalid-state", "vi", "127.0.0.1")
	if err != domain.ErrOAuthStateInvalid {
		t.Fatalf("expected ErrOAuthStateInvalid, got: %v", err)
	}
}

func TestLoginOrRegisterWithGitHub_ExchangeCodeError(t *testing.T) {
	service := setupFullService(&mockAuthRepo{}, &mockRoleRepo{}, &mockUserRepo{})
	state, _ := utils.GenerateOAuthState("supersecretkeyforunittestingjwttoken")

	ghMock := &mockGitHubClient{
		tokenErr: errors.New("exchange failed"),
	}
	service.SetGitHubClient(ghMock)
	service.SetOAuthRepo(&mockOAuthRepo{})

	_, err := service.LoginOrRegisterWithGitHub(context.Background(), "invalid-code", state, "vi", "127.0.0.1")
	if err == nil {
		t.Fatalf("expected error from exchange code failure")
	}
}

func TestLoginOrRegisterWithGitHub_NoVerifiedEmail(t *testing.T) {
	service := setupFullService(&mockAuthRepo{}, &mockRoleRepo{}, &mockUserRepo{})
	state, _ := utils.GenerateOAuthState("supersecretkeyforunittestingjwttoken")

	ghMock := &mockGitHubClient{
		userInfo: &oauth.GitHubUserInfo{
			ID:    999,
			Login: "noemailuser",
			Email: "", // No verified email
		},
	}
	service.SetGitHubClient(ghMock)
	service.SetOAuthRepo(&mockOAuthRepo{})

	_, err := service.LoginOrRegisterWithGitHub(context.Background(), "valid-code", state, "vi", "127.0.0.1")
	if err != domain.ErrOAuthEmailNotFound {
		t.Fatalf("expected ErrOAuthEmailNotFound, got: %v", err)
	}
}

func TestLoginOrRegisterWithGitHub_NewUserRegistration(t *testing.T) {
	roleRepo := &mockRoleRepo{
		roles: map[int]*domain.Role{
			2: {ID: 2, Name: "User", IsActive: true},
		},
	}
	userRepo := &mockUserRepo{
		usersByID:       make(map[int]*domain.User),
		usersByUsername: make(map[string]*domain.User),
		usersByEmail:    make(map[string]*domain.User),
	}
	authRepo := &mockAuthRepo{
		auths: make(map[string]*domain.Auth),
	}
	oauthRepo := &mockOAuthRepo{}

	service := setupFullService(authRepo, roleRepo, userRepo)
	state, _ := utils.GenerateOAuthState("supersecretkeyforunittestingjwttoken")

	ghMock := &mockGitHubClient{
		userInfo: &oauth.GitHubUserInfo{
			ID:        123456,
			Login:     "octocat",
			Name:      "The Octocat",
			Email:     "octocat@github.com",
			AvatarURL: "https://github.com/images/octocat.png",
		},
	}
	service.SetGitHubClient(ghMock)
	service.SetOAuthRepo(oauthRepo)

	// Since mockUserRepo.CreateWithAuth sets user.ID = 100, we mock authRepo accordingly
	// We'll pre-populate authRepo when CreateWithAuth is called or simulate auth entry:
	authRepo.auths["octocat"] = &domain.Auth{
		UserId:   100,
		Username: "octocat",
		RoleId:   2,
		IsActive: true,
	}

	resp, err := service.LoginOrRegisterWithGitHub(context.Background(), "code123", state, "vi", "127.0.0.1")
	if err != nil {
		t.Fatalf("expected nil error on registration, got: %v", err)
	}

	if resp == nil || resp.AccessToken == "" {
		t.Fatalf("expected non-empty access token")
	}
	if resp.RefreshToken == "" {
		t.Fatalf("expected non-empty refresh token")
	}
	if len(oauthRepo.accounts) != 1 {
		t.Errorf("expected 1 oauth account created, got: %d", len(oauthRepo.accounts))
	}
	if oauthRepo.accounts[0].Provider != "github" || oauthRepo.accounts[0].ProviderUserID != "123456" {
		t.Errorf("unexpected oauth account data: %+v", oauthRepo.accounts[0])
	}
}

func TestLoginOrRegisterWithGitHub_ExistingOAuthAccount(t *testing.T) {
	existingUser := &domain.User{
		ID:        10,
		Username:  "githubber",
		Email:     "githubber@example.com",
		FullName:  "GitHub User",
		IsActive:  true,
		IsDeleted: false,
	}
	existingAuth := &domain.Auth{
		UserId:   10,
		Username: "githubber",
		RoleId:   2,
		IsActive: true,
	}
	existingRole := &domain.Role{
		ID:       2,
		Name:     "User",
		IsActive: true,
	}

	roleRepo := &mockRoleRepo{roles: map[int]*domain.Role{2: existingRole}}
	userRepo := &mockUserRepo{usersByID: map[int]*domain.User{10: existingUser}}
	authRepo := &mockAuthRepo{auths: map[string]*domain.Auth{"githubber": existingAuth}}
	oauthRepo := &mockOAuthRepo{
		accounts: []*domain.OAuthAccount{
			{
				ID:             1,
				UserID:         10,
				Provider:       "github",
				ProviderUserID: "88888",
				Email:          "githubber@example.com",
			},
		},
	}

	service := setupFullService(authRepo, roleRepo, userRepo)
	state, _ := utils.GenerateOAuthState("supersecretkeyforunittestingjwttoken")

	ghMock := &mockGitHubClient{
		userInfo: &oauth.GitHubUserInfo{
			ID:    88888,
			Login: "githubber",
			Email: "githubber@example.com",
		},
	}
	service.SetGitHubClient(ghMock)
	service.SetOAuthRepo(oauthRepo)

	resp, err := service.LoginOrRegisterWithGitHub(context.Background(), "code123", state, "vi", "127.0.0.1")
	if err != nil {
		t.Fatalf("expected nil error on existing oauth account login, got: %v", err)
	}
	if resp == nil || resp.AccessToken == "" {
		t.Fatalf("expected non-empty access token")
	}
	if resp.User == nil || resp.User.ID != 10 {
		t.Errorf("expected user ID 10, got: %v", resp.User)
	}
}

func TestLoginOrRegisterWithGitHub_LinkExistingEmail(t *testing.T) {
	existingUser := &domain.User{
		ID:        15,
		Username:  "localuser",
		Email:     "shared@example.com",
		FullName:  "Local User",
		IsActive:  true,
		IsDeleted: false,
	}
	existingAuth := &domain.Auth{
		UserId:   15,
		Username: "localuser",
		RoleId:   2,
		IsActive: true,
	}
	role := &domain.Role{ID: 2, Name: "User", IsActive: true}

	roleRepo := &mockRoleRepo{roles: map[int]*domain.Role{2: role}}
	userRepo := &mockUserRepo{
		usersByID:    map[int]*domain.User{15: existingUser},
		usersByEmail: map[string]*domain.User{"shared@example.com": existingUser},
	}
	authRepo := &mockAuthRepo{auths: map[string]*domain.Auth{"localuser": existingAuth}}
	oauthRepo := &mockOAuthRepo{accounts: []*domain.OAuthAccount{}}

	service := setupFullService(authRepo, roleRepo, userRepo)
	state, _ := utils.GenerateOAuthState("supersecretkeyforunittestingjwttoken")

	ghMock := &mockGitHubClient{
		userInfo: &oauth.GitHubUserInfo{
			ID:    77777,
			Login: "ghuser",
			Email: "shared@example.com",
		},
	}
	service.SetGitHubClient(ghMock)
	service.SetOAuthRepo(oauthRepo)

	resp, err := service.LoginOrRegisterWithGitHub(context.Background(), "code123", state, "vi", "127.0.0.1")
	if err != nil {
		t.Fatalf("expected nil error on link existing email, got: %v", err)
	}
	if resp == nil || resp.User.ID != 15 {
		t.Errorf("expected linked user ID 15, got: %v", resp.User)
	}
	if len(oauthRepo.accounts) != 1 {
		t.Errorf("expected 1 oauth account created for linking, got: %d", len(oauthRepo.accounts))
	}
	if oauthRepo.accounts[0].UserID != 15 {
		t.Errorf("expected oauth account linked to user 15, got: %d", oauthRepo.accounts[0].UserID)
	}
	if oauthRepo.accounts[0].AccessToken == "" {
		t.Errorf("expected encrypted access token to be saved")
	}
	decrypted, err := utils.DecryptAESGCM(oauthRepo.accounts[0].AccessToken, "supersecretkeyforunittestingjwttoken")
	if err != nil || decrypted != "mock-access-token" {
		t.Errorf("expected decrypted token 'mock-access-token', got: %s (err: %v)", decrypted, err)
	}
}

func TestGetGitHubRepositories_Success(t *testing.T) {
	service := setupFullService(&mockAuthRepo{}, &mockRoleRepo{}, &mockUserRepo{})
	encryptedToken, _ := utils.EncryptAESGCM("valid-token", "supersecretkeyforunittestingjwttoken")

	oauthRepo := &mockOAuthRepo{
		accounts: []*domain.OAuthAccount{
			{
				ID:             1,
				UserID:         10,
				Provider:       "github",
				ProviderUserID: "12345",
				AccessToken:    encryptedToken,
			},
		},
	}

	ghMock := &mockGitHubClient{
		repos: []*domain.GitHubRepository{
			{
				ID:       101,
				Name:     "backend-business-chat",
				FullName: "test/backend-business-chat",
				Private:  true,
			},
		},
	}

	service.SetOAuthRepo(oauthRepo)
	service.SetGitHubClient(ghMock)

	repos, err := service.GetGitHubRepositories(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(repos) != 1 || repos[0].Name != "backend-business-chat" {
		t.Fatalf("unexpected repos returned: %v", repos)
	}
}

func TestGetGitHubRepositories_NotLinked(t *testing.T) {
	service := setupFullService(&mockAuthRepo{}, &mockRoleRepo{}, &mockUserRepo{})
	oauthRepo := &mockOAuthRepo{accounts: []*domain.OAuthAccount{}}
	ghMock := &mockGitHubClient{}

	service.SetOAuthRepo(oauthRepo)
	service.SetGitHubClient(ghMock)

	_, err := service.GetGitHubRepositories(context.Background(), 99)
	if !errors.Is(err, domain.ErrGitHubAccountNotLinked) {
		t.Fatalf("expected ErrGitHubAccountNotLinked, got: %v", err)
	}
}

func TestCreateGitHubIssue_Success(t *testing.T) {
	service := setupFullService(&mockAuthRepo{}, &mockRoleRepo{}, &mockUserRepo{})
	encryptedToken, _ := utils.EncryptAESGCM("valid-token", "supersecretkeyforunittestingjwttoken")

	oauthRepo := &mockOAuthRepo{
		accounts: []*domain.OAuthAccount{
			{
				ID:             1,
				UserID:         10,
				Provider:       "github",
				ProviderUserID: "12345",
				AccessToken:    encryptedToken,
			},
		},
	}

	ghMock := &mockGitHubClient{
		issueResp: &domain.GitHubIssueResponse{
			ID:     999,
			Number: 42,
			Title:  "Bug: login issue",
			State:  "open",
		},
	}

	service.SetOAuthRepo(oauthRepo)
	service.SetGitHubClient(ghMock)

	req := domain.GitHubCreateIssueRequest{
		Title: "Bug: login issue",
		Body:  "Detailed bug description",
	}

	issue, err := service.CreateGitHubIssue(context.Background(), 10, "owner", "repo", req)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if issue == nil || issue.Number != 42 {
		t.Fatalf("unexpected issue returned: %v", issue)
	}
}

func TestCreateGitHubIssue_NotLinked(t *testing.T) {
	service := setupFullService(&mockAuthRepo{}, &mockRoleRepo{}, &mockUserRepo{})
	oauthRepo := &mockOAuthRepo{accounts: []*domain.OAuthAccount{}}
	ghMock := &mockGitHubClient{}

	service.SetOAuthRepo(oauthRepo)
	service.SetGitHubClient(ghMock)

	req := domain.GitHubCreateIssueRequest{Title: "Issue"}
	_, err := service.CreateGitHubIssue(context.Background(), 99, "owner", "repo", req)
	if !errors.Is(err, domain.ErrGitHubAccountNotLinked) {
		t.Fatalf("expected ErrGitHubAccountNotLinked, got: %v", err)
	}
}



