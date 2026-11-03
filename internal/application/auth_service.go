package application

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sixgillkrahs/backend-business-chat/internal/application/dto"
	"github.com/sixgillkrahs/backend-business-chat/internal/config"
	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
	"github.com/sixgillkrahs/backend-business-chat/pkg/utils"
)

type AuthService struct {
	config         *config.Config
	actionRepo     domain.ActionRepository
	authRepo       domain.AuthRepository
	resourcesRepo  domain.ResourceRepository
	policyRepo     domain.PolicyRepository
	permissionRepo domain.PermissionRepository
	roleRepo       domain.RoleRepository
	userRepo       domain.UserRepository
}

func NewAuthService(
	cfg *config.Config,
	actionRepo domain.ActionRepository,
	authRepo domain.AuthRepository,
	resourcesRepo domain.ResourceRepository,
	policyRepo domain.PolicyRepository,
	permissionRepo domain.PermissionRepository,
	roleRepo domain.RoleRepository,
	userRepo domain.UserRepository,
) *AuthService {
	return &AuthService{
		config:         cfg,
		actionRepo:     actionRepo,
		authRepo:       authRepo,
		resourcesRepo:  resourcesRepo,
		policyRepo:     policyRepo,
		permissionRepo: permissionRepo,
		roleRepo:       roleRepo,
		userRepo:       userRepo,
	}
}

func (s *AuthService) GetAllActions(ctx context.Context) ([]domain.Action, error) {
	return s.actionRepo.GetAllActions(ctx)
}

func (s *AuthService) GetAllResources(ctx context.Context) ([]domain.Resource, error) {
	return s.resourcesRepo.GetAllResources(ctx)
}

func (s *AuthService) InitDefaultResources(ctx context.Context) error {
	for _, res := range domain.DefaultResources {
		err := s.resourcesRepo.Upsert(ctx, &res)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *AuthService) GetPoliciesPaging(ctx context.Context, page, limit int, search string) (*dto.PolicyPaginationResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	policies, err := s.policyRepo.GetPoliciesPaging(ctx, search, offset, limit)
	if err != nil {
		return nil, err
	}

	total, err := s.policyRepo.Count(ctx, search)
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &dto.PolicyPaginationResponse{
		Items:      policies,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *AuthService) CountPolicies(ctx context.Context, search string) (int64, error) {
	return s.policyRepo.Count(ctx, search)
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest, lang string, ip string) (string, error) {
	auth, err := s.authRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", err
	}

	if !utils.CheckPasswordHash(req.Password, auth.PasswordHash) {
		return "", domain.ErrInvalidCredentials
	}

	if !auth.IsActive {
		return "", domain.ErrUserLocked
	}
	duration := time.Duration(s.config.Server.Jwt.Duration) * time.Minute
	accessToken, err := utils.GenerateToken(int64(auth.UserId), auth.Username, strconv.Itoa(auth.RoleId), ip, s.config.Server.Jwt.SecretKey, duration)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

func (s *AuthService) CreatePolicy(ctx context.Context, req dto.PolicyRequest, lang string) (string, error) {
	action, err := s.actionRepo.GetActionById(ctx, req.ActionId)
	if err != nil {
		return "", err
	}
	resource, err := s.resourcesRepo.GetResourceById(ctx, req.ResourceId)
	if err != nil {
		return "", err
	}

	policy := &domain.Policy{
		Name:        req.Name,
		Description: req.Description,
		ActionID:    action.ID,
		ResourceID:  resource.ID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.policyRepo.Insert(ctx, policy); err != nil {
		return "", err
	}

	return "", nil
}

func (s *AuthService) GetPolicyById(ctx context.Context, id int, lang string) (domain.Policy, error) {
	policy, err := s.policyRepo.GetPolicyById(ctx, id)
	if err != nil {
		return domain.Policy{}, err
	}
	return policy, nil
}

func (s *AuthService) UpdatePolicyById(ctx context.Context, id int, req dto.UpdatePolicyRequest, lang string) (domain.Policy, error) {
	policy, err := s.policyRepo.GetPolicyById(ctx, id)
	if err != nil {
		return domain.Policy{}, err
	}
	if req.Name != nil {
		policy.Name = *req.Name
	}
	if req.Description != nil {
		policy.Description = *req.Description
	}
	if req.ResourceId != nil {
		policy.ResourceID = *req.ResourceId
	}
	if req.ActionId != nil {
		policy.ActionID = *req.ActionId
	}
	policy.UpdatedAt = time.Now()
	if err := s.policyRepo.Update(ctx, &policy); err != nil {
		return domain.Policy{}, err
	}
	return policy, nil
}

func (s *AuthService) DeletePolicyById(ctx context.Context, id int, lang string) error {
	_, err := s.policyRepo.GetPolicyById(ctx, id)
	if err != nil {
		return err
	}
	return s.policyRepo.Delete(ctx, id)
}

func (s *AuthService) CreatePermission(ctx context.Context, req dto.CreatePermissionRequest, lang string) (int, error) {
	name := strings.TrimSpace(req.Name)
	desc := strings.TrimSpace(req.Description)

	policyMap := make(map[int]struct{})
	var uniquePolicyIDs []int
	for _, id := range req.PolicyIDs {
		if _, exists := policyMap[id]; !exists {
			policyMap[id] = struct{}{}
			uniquePolicyIDs = append(uniquePolicyIDs, id)
		}
	}

	if len(uniquePolicyIDs) == 0 {
		return 0, domain.ErrEmptyPolicies
	}

	existing, err := s.permissionRepo.GetPermissionByName(ctx, name)
	if err != nil {
		return 0, err
	}
	if existing != nil {
		return existing.ID, domain.ErrPermissionNameDuplicate
	}

	valid, err := s.permissionRepo.ValidatePolicyIDs(ctx, uniquePolicyIDs)
	if err != nil {
		return 0, err
	}
	if !valid {
		return 0, domain.ErrPolicyNotFound
	}

	now := time.Now()
	perm := &domain.Permission{
		Name:        name,
		Description: desc,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.permissionRepo.CreateWithPolicies(ctx, perm, uniquePolicyIDs); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, domain.ErrPermissionNameDuplicate
		}
		return 0, err
	}

	return perm.ID, nil
}

func (s *AuthService) GetPermissionsPaging(ctx context.Context, limit, offset int, search string) (*dto.PermissionPaginationResponse, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	search = strings.TrimSpace(search)
	if len(search) > 100 {
		search = search[:100]
	}

	permissions, err := s.permissionRepo.GetPermissionsPaging(ctx, search, offset, limit)
	if err != nil {
		return nil, err
	}

	total, err := s.permissionRepo.Count(ctx, search)
	if err != nil {
		return nil, err
	}

	return &dto.PermissionPaginationResponse{
		Items:  permissions,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *AuthService) GetPermissionById(ctx context.Context, id int) (*domain.Permission, error) {
	if id <= 0 {
		return nil, domain.ErrPermissionNotFound
	}

	perm, err := s.permissionRepo.GetPermissionById(ctx, id)
	if err != nil {
		return nil, err
	}
	if perm == nil {
		return nil, domain.ErrPermissionNotFound
	}

	return perm, nil
}

func (s *AuthService) UpdatePermissionById(ctx context.Context, id int, req dto.UpdatePermissionRequest, lang string) (*domain.Permission, error) {
	if id <= 0 {
		return nil, domain.ErrPermissionNotFound
	}

	perm, err := s.permissionRepo.GetPermissionById(ctx, id)
	if err != nil {
		return nil, err
	}
	if perm == nil {
		return nil, domain.ErrPermissionNotFound
	}

	if req.Name == nil && req.Description == nil && req.PolicyIDs == nil {
		return perm, nil
	}

	if req.Name != nil {
		trimmedName := strings.TrimSpace(*req.Name)
		if len(trimmedName) < 3 || len(trimmedName) > 50 {
			return nil, errors.New("tên permission phải từ 3 đến 50 ký tự")
		}
		if trimmedName != perm.Name {
			existing, err := s.permissionRepo.GetPermissionByName(ctx, trimmedName)
			if err != nil {
				return nil, err
			}
			if existing != nil && existing.ID != id {
				return nil, domain.ErrPermissionNameDuplicate
			}
			perm.Name = trimmedName
		}
	}

	if req.Description != nil {
		trimmedDesc := strings.TrimSpace(*req.Description)
		if len(trimmedDesc) > 255 {
			return nil, errors.New("mô tả không được vượt quá 255 ký tự")
		}
		perm.Description = trimmedDesc
	}

	var uniquePolicyIDs []int
	if req.PolicyIDs != nil {
		policyMap := make(map[int]struct{})
		for _, pid := range req.PolicyIDs {
			if pid <= 0 {
				return nil, domain.ErrPolicyNotFound
			}
			if _, exists := policyMap[pid]; !exists {
				policyMap[pid] = struct{}{}
				uniquePolicyIDs = append(uniquePolicyIDs, pid)
			}
		}

		if len(uniquePolicyIDs) == 0 {
			return nil, domain.ErrEmptyPolicies
		}

		valid, err := s.permissionRepo.ValidatePolicyIDs(ctx, uniquePolicyIDs)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, domain.ErrPolicyNotFound
		}
	}

	perm.UpdatedAt = time.Now()

	if err := s.permissionRepo.UpdateWithPolicies(ctx, perm, uniquePolicyIDs); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrPermissionNameDuplicate
		}
		return nil, err
	}

	return s.permissionRepo.GetPermissionById(ctx, id)
}

func (s *AuthService) DeletePermissionById(ctx context.Context, id int, lang string) error {
	if id <= 0 {
		return domain.ErrPermissionNotFound
	}

	perm, err := s.permissionRepo.GetPermissionById(ctx, id)
	if err != nil {
		return err
	}
	if perm == nil {
		return domain.ErrPermissionNotFound
	}

	return s.permissionRepo.DeletePermissionById(ctx, id)
}

func (s *AuthService) CreateRole(ctx context.Context, req dto.CreateRoleRequest, lang string) (int, error) {
	name := strings.TrimSpace(req.Name)
	desc := strings.TrimSpace(req.Description)

	if len(name) < 3 || len(name) > 50 {
		return 0, errors.New("tên vai trò phải từ 3 đến 50 ký tự")
	}

	permMap := make(map[int]struct{})
	var uniquePermIDs []int
	for _, id := range req.PermissionIDs {
		if id <= 0 {
			return 0, domain.ErrPermissionNotFound
		}
		if _, exists := permMap[id]; !exists {
			permMap[id] = struct{}{}
			uniquePermIDs = append(uniquePermIDs, id)
		}
	}

	if len(uniquePermIDs) == 0 {
		return 0, domain.ErrEmptyRolePermissions
	}

	existing, err := s.roleRepo.GetRoleByName(ctx, name)
	if err != nil {
		return 0, err
	}
	if existing != nil {
		return existing.ID, domain.ErrRoleNameDuplicate
	}

	valid, err := s.roleRepo.ValidatePermissionIDs(ctx, uniquePermIDs)
	if err != nil {
		return 0, err
	}
	if !valid {
		return 0, domain.ErrPermissionNotFound
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	now := time.Now()
	role := &domain.Role{
		Name:        name,
		Description: desc,
		IsActive:    isActive,
		IsSystem:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.roleRepo.CreateWithPermissions(ctx, role, uniquePermIDs); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, domain.ErrRoleNameDuplicate
		}
		return 0, err
	}

	return role.ID, nil
}

func (s *AuthService) GetRolesPaging(ctx context.Context, page, limit int, search string) (*dto.RolePaginationResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	search = strings.TrimSpace(search)
	if len(search) > 100 {
		search = search[:100]
	}

	offset := (page - 1) * limit

	roles, err := s.roleRepo.GetRolesPaging(ctx, search, offset, limit)
	if err != nil {
		return nil, err
	}

	total, err := s.roleRepo.Count(ctx, search)
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &dto.RolePaginationResponse{
		Items:      roles,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *AuthService) UpdateRoleById(ctx context.Context, id int, req dto.UpdateRoleRequest, lang string) (*domain.Role, error) {
	if id <= 0 {
		return nil, domain.ErrRoleNotFound
	}

	role, err := s.roleRepo.GetRoleById(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, domain.ErrRoleNotFound
	}

	if role.IsSystem {
		return nil, domain.ErrSystemRoleProtected
	}

	if req.Name == nil && req.Description == nil && req.IsActive == nil && req.PermissionIDs == nil {
		return role, nil
	}

	if req.Name != nil {
		trimmedName := strings.TrimSpace(*req.Name)
		if len(trimmedName) < 3 || len(trimmedName) > 50 {
			return nil, errors.New("tên vai trò phải từ 3 đến 50 ký tự")
		}
		if trimmedName != role.Name {
			existing, err := s.roleRepo.GetRoleByName(ctx, trimmedName)
			if err != nil {
				return nil, err
			}
			if existing != nil && existing.ID != id {
				return nil, domain.ErrRoleNameDuplicate
			}
			role.Name = trimmedName
		}
	}

	if req.Description != nil {
		trimmedDesc := strings.TrimSpace(*req.Description)
		if len(trimmedDesc) > 255 {
			return nil, errors.New("mô tả không được vượt quá 255 ký tự")
		}
		role.Description = trimmedDesc
	}

	if req.IsActive != nil {
		role.IsActive = *req.IsActive
	}

	var uniquePermIDs []int
	if req.PermissionIDs != nil {
		permMap := make(map[int]struct{})
		for _, pid := range req.PermissionIDs {
			if pid <= 0 {
				return nil, domain.ErrPermissionNotFound
			}
			if _, exists := permMap[pid]; !exists {
				permMap[pid] = struct{}{}
				uniquePermIDs = append(uniquePermIDs, pid)
			}
		}

		if len(uniquePermIDs) == 0 {
			return nil, domain.ErrEmptyRolePermissions
		}

		valid, err := s.roleRepo.ValidatePermissionIDs(ctx, uniquePermIDs)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, domain.ErrPermissionNotFound
		}
	}

	role.UpdatedAt = time.Now()

	if err := s.roleRepo.UpdateWithPermissions(ctx, role, uniquePermIDs); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrRoleNameDuplicate
		}
		return nil, err
	}

	return s.roleRepo.GetRoleById(ctx, id)
}

func (s *AuthService) DeleteRoleById(ctx context.Context, id int, lang string) error {
	if id <= 0 {
		return domain.ErrRoleNotFound
	}

	role, err := s.roleRepo.GetRoleById(ctx, id)
	if err != nil {
		return err
	}
	if role == nil {
		return domain.ErrRoleNotFound
	}

	if role.IsSystem {
		return domain.ErrSystemRoleCannotDelete
	}

	userCount, err := s.roleRepo.CountUsersByRoleId(ctx, id)
	if err != nil {
		return err
	}
	if userCount > 0 {
		return domain.ErrRoleInUse
	}

	return s.roleRepo.DeleteRoleById(ctx, id)
}

func (s *AuthService) CreateUser(ctx context.Context, req dto.CreateUserRequest, lang string, createdBy string) (*domain.User, error) {
	username := strings.TrimSpace(req.Username)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	fullName := strings.TrimSpace(req.FullName)

	if len(req.Password) < 8 || len(req.Password) > 72 {
		return nil, domain.ErrPasswordTooShort
	}

	if createdBy == "" {
		createdBy = "system"
	} else if len(createdBy) > 20 {
		createdBy = createdBy[:20]
	}

	role, err := s.roleRepo.GetRoleById(ctx, req.RoleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, domain.ErrRoleNotFound
	}

	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, domain.ErrUsernameDuplicate
	}

	existingEmail, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingEmail != nil {
		return nil, domain.ErrEmailDuplicate
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username:    username,
		Email:       email,
		FullName:    fullName,
		AvatarURL:   strings.TrimSpace(req.AvatarURL),
		PhoneNumber: strings.TrimSpace(req.PhoneNumber),
		Address:     strings.TrimSpace(req.Address),
		RoleID:      &req.RoleID,
		IsActive:    true,
		IsDeleted:   false,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := s.userRepo.CreateWithAuth(ctx, user, passwordHash); err != nil {
		return nil, err
	}

	user.Role = role
	return user, nil
}
