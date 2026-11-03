package application_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/sixgillkrahs/backend-business-chat/internal/application"
	"github.com/sixgillkrahs/backend-business-chat/internal/application/dto"
	"github.com/sixgillkrahs/backend-business-chat/internal/config"
	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
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
	return nil, nil
}
func (m *mockRoleRepo) CreateWithPermissions(ctx context.Context, role *domain.Role, permissionIDs []int) error {
	return nil
}
func (m *mockRoleRepo) GetRolesPaging(ctx context.Context, search string, offset, limit int) ([]domain.Role, error) {
	return nil, nil
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

// Mock User Repository
type mockUserRepo struct {
	users           []domain.User
	usersByUsername map[string]*domain.User
	usersByEmail    map[string]*domain.User
	createdUser     *domain.User
	createdHash     string
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
func (m *mockUserRepo) FindByID(ctx context.Context, id int) (*domain.User, error) { return nil, nil }
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
func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error { return nil }
func (m *mockUserRepo) Delete(ctx context.Context, id int) error            { return nil }
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

func setupService(roleRepo domain.RoleRepository, userRepo domain.UserRepository) *application.AuthService {
	cfg := &config.Config{}
	return application.NewAuthService(cfg, nil, nil, nil, nil, nil, roleRepo, userRepo)
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

	// Verify password hash was hashed with bcrypt
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
		Password: "123", // too short (< 8 chars)
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

	// Call with 0 limit and 0 offset -> should default to limit 10, offset 0
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

	// 1. Limit > 50 clamped to 50
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
