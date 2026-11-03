package application_test

import (
	"context"
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
	usersByUsername map[string]*domain.User
	usersByEmail    map[string]*domain.User
	createdUser     *domain.User
	createdHash     string
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
	return nil, nil
}
func (m *mockUserRepo) Count(ctx context.Context, search string) (int64, error) { return 0, nil }

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
