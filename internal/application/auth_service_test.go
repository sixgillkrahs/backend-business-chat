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
