package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sixgillkrahs/backend-business-chat/internal/application/dto"
	"github.com/sixgillkrahs/backend-business-chat/internal/config"
	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/oauth"
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
	oauthRepo      domain.OAuthRepository
	gitHubClient   oauth.GitHubClient
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
	extra ...interface{},
) *AuthService {
	s := &AuthService{
		config:         cfg,
		actionRepo:     actionRepo,
		authRepo:       authRepo,
		resourcesRepo:  resourcesRepo,
		policyRepo:     policyRepo,
		permissionRepo: permissionRepo,
		roleRepo:       roleRepo,
		userRepo:       userRepo,
	}

	for _, opt := range extra {
		switch v := opt.(type) {
		case domain.OAuthRepository:
			s.oauthRepo = v
		case oauth.GitHubClient:
			s.gitHubClient = v
		}
	}

	if s.gitHubClient == nil && cfg != nil {
		s.gitHubClient = oauth.NewGitHubClient(cfg.OAuth.GitHub)
	}

	return s
}

func (s *AuthService) SetOAuthRepo(repo domain.OAuthRepository) {
	s.oauthRepo = repo
}

func (s *AuthService) SetGitHubClient(client oauth.GitHubClient) {
	s.gitHubClient = client
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

func (s *AuthService) GetConfig() *config.Config {
	return s.config
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest, lang string, ip string) (*dto.LoginResponse, error) {
	username := strings.TrimSpace(req.Username)
	password := req.Password

	const dummyBcryptHash = "$2a$10$7EqJtq98hPqEX7fNZaFWoO.8/k0a5d4gB9eG4eX5YvG4b.8s4Z0K6"

	if len(username) < 3 || len(username) > 50 || len(password) < 6 || len(password) > 72 {
		_ = utils.CheckPasswordHash(password, dummyBcryptHash)
		return nil, domain.ErrInvalidCredentials
	}

	auth, err := s.authRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			_ = utils.CheckPasswordHash(password, dummyBcryptHash)
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if !utils.CheckPasswordHash(password, auth.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	if !auth.IsActive || auth.IsDeleted {
		return nil, domain.ErrUserLocked
	}

	var user *domain.User
	if s.userRepo != nil && auth.UserId > 0 {
		var err error
		user, err = s.userRepo.FindByID(ctx, auth.UserId)
		if err != nil {
			return nil, err
		}
		if user == nil || user.IsDeleted {
			return nil, domain.ErrInvalidCredentials
		}
		if !user.IsActive {
			return nil, domain.ErrUserLocked
		}
	}

	return s.generateLoginResponse(ctx, &auth, user, ip)
}

func (s *AuthService) RefreshToken(ctx context.Context, tokenStr string, lang string, ip string) (*dto.LoginResponse, error) {
	tokenStr = strings.TrimSpace(tokenStr)
	if tokenStr == "" {
		return nil, domain.ErrInvalidRefreshToken
	}

	secretKey := "secret"
	if s.config != nil && s.config.Server.Jwt.SecretKey != "" {
		secretKey = s.config.Server.Jwt.SecretKey
	}

	claims, err := utils.ValidateRefreshToken(tokenStr, secretKey)
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	auth, err := s.authRepo.FindByUsername(ctx, claims.Username)
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	if !auth.IsActive || auth.IsDeleted {
		return nil, domain.ErrUserLocked
	}

	var user *domain.User
	if s.userRepo != nil && auth.UserId > 0 {
		var err error
		user, err = s.userRepo.FindByID(ctx, auth.UserId)
		if err != nil {
			return nil, err
		}
		if user == nil || user.IsDeleted {
			return nil, domain.ErrInvalidRefreshToken
		}
		if !user.IsActive {
			return nil, domain.ErrUserLocked
		}
	}

	return s.generateLoginResponse(ctx, &auth, user, ip)
}

func (s *AuthService) generateLoginResponse(ctx context.Context, auth *domain.Auth, user *domain.User, ip string) (*dto.LoginResponse, error) {
	var userResp *dto.UserLoginResponse
	var permissions []domain.Permission
	var roleResp *dto.RoleResponse

	if user != nil {
		userResp = &dto.UserLoginResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			AvatarURL: user.AvatarURL,
		}
	}

	roleClaim := strconv.Itoa(auth.RoleId)

	if s.roleRepo != nil && auth.RoleId > 0 {
		role, err := s.roleRepo.GetRoleById(ctx, auth.RoleId)
		if err != nil {
			return nil, err
		}
		if role != nil {
			if !role.IsActive {
				return nil, domain.ErrRoleLocked
			}
			roleResp = &dto.RoleResponse{
				ID:          role.ID,
				Name:        role.Name,
				Description: role.Description,
				IsActive:    role.IsActive,
				IsSystem:    role.IsSystem,
			}
			permissions = role.Permissions
		}
	}

	if userResp != nil {
		userResp.Role = roleResp
		userResp.Permissions = permissions
	}

	duration := 15 * time.Minute
	if s.config != nil && s.config.Server.Jwt.Duration > 0 {
		duration = time.Duration(s.config.Server.Jwt.Duration) * time.Minute
	}

	secretKey := "secret"
	if s.config != nil && s.config.Server.Jwt.SecretKey != "" {
		secretKey = s.config.Server.Jwt.SecretKey
	}

	accessToken, err := utils.GenerateAccessToken(int64(auth.UserId), auth.Username, roleClaim, ip, secretKey, duration)
	if err != nil {
		return nil, err
	}

	refreshDuration := 7 * 24 * time.Hour
	refreshToken, err := utils.GenerateRefreshToken(int64(auth.UserId), auth.Username, roleClaim, ip, secretKey, refreshDuration)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(duration.Seconds()),
		User:         userResp,
	}, nil
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

func (s *AuthService) UpdateRoleStatus(ctx context.Context, id int, req dto.UpdateRoleStatusRequest, lang string) (*domain.Role, error) {
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

	targetStatus := !role.IsActive
	if req.IsActive != nil {
		targetStatus = *req.IsActive
	}

	if err := s.roleRepo.UpdateRoleStatus(ctx, id, targetStatus); err != nil {
		return nil, err
	}

	return s.roleRepo.GetRoleById(ctx, id)
}

func (s *AuthService) DeactivateRoleById(ctx context.Context, id int, lang string) error {
	isActive := false
	_, err := s.UpdateRoleStatus(ctx, id, dto.UpdateRoleStatusRequest{IsActive: &isActive}, lang)
	if errors.Is(err, domain.ErrSystemRoleProtected) {
		return domain.ErrSystemRoleCannotDeactivate
	}
	return err
}

func (s *AuthService) getDefaultRole(ctx context.Context) (*domain.Role, error) {
	if s.roleRepo == nil {
		return nil, domain.ErrRoleNotFound
	}

	for _, name := range []string{"User", "USER", "Member", "ROLE_USER"} {
		role, err := s.roleRepo.GetRoleByName(ctx, name)
		if err == nil && role != nil && role.IsActive {
			return role, nil
		}
	}

	roles, err := s.roleRepo.GetRolesPaging(ctx, "", 0, 50)
	if err == nil {
		for _, r := range roles {
			if r.IsActive && !r.IsSystem {
				return &r, nil
			}
		}
		for _, r := range roles {
			if r.IsActive {
				return &r, nil
			}
		}
	}

	return nil, domain.ErrRoleNotFound
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest, lang string) (*domain.User, error) {
	username := strings.TrimSpace(req.Username)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	fullName := strings.TrimSpace(req.FullName)

	if len(req.Password) < 8 || len(req.Password) > 72 {
		return nil, domain.ErrPasswordTooShort
	}
	if !utils.IsPasswordComplex(req.Password) {
		return nil, domain.ErrPasswordTooWeak
	}

	role, err := s.getDefaultRole(ctx)
	if err != nil {
		return nil, err
	}
	if role == nil || !role.IsActive {
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
		RoleID:      &role.ID,
		IsActive:    true,
		IsDeleted:   false,
		CreatedBy:   "register",
		UpdatedBy:   "register",
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

func (s *AuthService) CreateUser(ctx context.Context, req dto.CreateUserRequest, lang string, createdBy string) (*domain.User, error) {
	username := strings.TrimSpace(req.Username)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	fullName := strings.TrimSpace(req.FullName)

	if len(req.Password) < 8 || len(req.Password) > 72 {
		return nil, domain.ErrPasswordTooShort
	}
	if !utils.IsPasswordComplex(req.Password) {
		return nil, domain.ErrPasswordTooWeak
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

func (s *AuthService) GetUsersPaging(ctx context.Context, limit, offset int, search string) (*dto.UserPaginationResponse, error) {
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

	users, err := s.userRepo.GetUsersPaging(ctx, search, offset, limit)
	if err != nil {
		return nil, err
	}

	total, err := s.userRepo.Count(ctx, search)
	if err != nil {
		return nil, err
	}

	return &dto.UserPaginationResponse{
		Items:  users,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}
func (s *AuthService) GetUserById(ctx context.Context, id int) (*domain.User, error) {
	if id <= 0 {
		return nil, domain.ErrUserNotFound
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	return user, nil
}

func (s *AuthService) UpdateUserById(ctx context.Context, id int, req dto.UpdateUserRequest, lang string, updatedBy string) (*domain.User, error) {
	if id <= 0 {
		return nil, domain.ErrUserNotFound
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	if updatedBy == "" {
		updatedBy = "system"
	} else if len(updatedBy) > 20 {
		updatedBy = updatedBy[:20]
	}

	if req.Username != nil {
		trimmedUsername := strings.TrimSpace(*req.Username)
		if trimmedUsername != user.Username {
			existing, err := s.userRepo.FindByUsername(ctx, trimmedUsername)
			if err != nil {
				return nil, err
			}
			if existing != nil && existing.ID != user.ID {
				return nil, domain.ErrUsernameDuplicate
			}
			user.Username = trimmedUsername
		}
	}

	if req.Email != nil {
		trimmedEmail := strings.ToLower(strings.TrimSpace(*req.Email))
		if trimmedEmail != user.Email {
			existing, err := s.userRepo.FindByEmail(ctx, trimmedEmail)
			if err != nil {
				return nil, err
			}
			if existing != nil && existing.ID != user.ID {
				return nil, domain.ErrEmailDuplicate
			}
			user.Email = trimmedEmail
		}
	}

	if req.FullName != nil {
		trimmedFullName := strings.TrimSpace(*req.FullName)
		user.FullName = trimmedFullName
	}

	if req.PhoneNumber != nil {
		trimmedPhone := strings.TrimSpace(*req.PhoneNumber)
		if len(trimmedPhone) > 20 {
			return nil, errors.New("số điện thoại tối đa 20 ký tự")
		}
		user.PhoneNumber = trimmedPhone
	}

	if req.Address != nil {
		trimmedAddress := strings.TrimSpace(*req.Address)
		if len(trimmedAddress) > 255 {
			return nil, errors.New("địa chỉ tối đa 255 ký tự")
		}
		user.Address = trimmedAddress
	}

	if req.AvatarURL != nil {
		trimmedAvatar := strings.TrimSpace(*req.AvatarURL)
		if len(trimmedAvatar) > 255 {
			return nil, errors.New("đường dẫn ảnh đại diện tối đa 255 ký tự")
		}
		user.AvatarURL = trimmedAvatar
	}

	var role *domain.Role
	if req.RoleID != nil {
		if *req.RoleID <= 0 {
			return nil, domain.ErrRoleNotFound
		}
		var err error
		role, err = s.roleRepo.GetRoleById(ctx, *req.RoleID)
		if err != nil {
			return nil, err
		}
		if role == nil {
			return nil, domain.ErrRoleNotFound
		}
		user.RoleID = req.RoleID
	} else if user.RoleID != nil {
		var err error
		role, err = s.roleRepo.GetRoleById(ctx, *user.RoleID)
		if err != nil {
			return nil, err
		}
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	user.UpdatedBy = updatedBy
	user.UpdatedAt = time.Now()

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	user.Role = role
	return user, nil
}

func (s *AuthService) DeleteUserById(ctx context.Context, id int, lang string, deletedBy string) error {
	if id <= 0 {
		return domain.ErrUserNotFound
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	// OWASP A01: Prevent deleting system administrators
	if user.Username == "admin" || (user.Role != nil && user.Role.IsSystem) {
		return domain.ErrSystemUserCannotDelete
	}

	// OWASP A01: Prevent deleting own account to avoid lockout
	if deletedBy != "" && strings.EqualFold(user.Username, deletedBy) {
		return domain.ErrCannotDeleteSelf
	}

	if deletedBy == "" {
		deletedBy = "system"
	} else if len(deletedBy) > 20 {
		deletedBy = deletedBy[:20]
	}

	return s.userRepo.Delete(ctx, id, deletedBy)
}

func (s *AuthService) GetGitHubAuthURL(ctx context.Context) (string, string, error) {
	if s.gitHubClient == nil {
		return "", "", errors.New("github oauth client is not initialized")
	}

	secretKey := "oauth-secret-key"
	if s.config != nil && s.config.Server.Jwt.SecretKey != "" {
		secretKey = s.config.Server.Jwt.SecretKey
	}

	state, err := utils.GenerateOAuthState(secretKey)
	if err != nil {
		return "", "", err
	}

	authURL := s.gitHubClient.GetAuthURL(state)
	return authURL, state, nil
}

func (s *AuthService) LoginOrRegisterWithGitHub(ctx context.Context, code string, state string, lang string, ip string) (*dto.LoginResponse, error) {
	code = strings.TrimSpace(code)
	state = strings.TrimSpace(state)
	if code == "" || state == "" {
		return nil, domain.ErrOAuthStateInvalid
	}

	secretKey := "oauth-secret-key"
	if s.config != nil && s.config.Server.Jwt.SecretKey != "" {
		secretKey = s.config.Server.Jwt.SecretKey
	}

	if !utils.ValidateOAuthState(state, secretKey, 10*time.Minute) {
		return nil, domain.ErrOAuthStateInvalid
	}

	if s.gitHubClient == nil {
		return nil, errors.New("github oauth client is not initialized")
	}
	if s.oauthRepo == nil {
		return nil, errors.New("oauth repository is not initialized")
	}

	ghToken, err := s.gitHubClient.ExchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	encryptedToken, err := utils.EncryptAESGCM(ghToken, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt oauth access token: %w", err)
	}

	ghUser, err := s.gitHubClient.GetUserInfo(ctx, ghToken)
	if err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(ghUser.Email))
	if email == "" {
		return nil, domain.ErrOAuthEmailNotFound
	}

	ghIDStr := strconv.FormatInt(ghUser.ID, 10)

	oauthAcc, err := s.oauthRepo.FindByProviderAndUserID(ctx, "github", ghIDStr)
	if err != nil {
		return nil, err
	}

	if oauthAcc != nil {
		_ = s.oauthRepo.UpdateAccessToken(ctx, oauthAcc.ID, encryptedToken)

		user, err := s.userRepo.FindByID(ctx, oauthAcc.UserID)
		if err != nil {
			return nil, err
		}
		if user == nil || user.IsDeleted || !user.IsActive {
			return nil, domain.ErrUserLocked
		}

		auth, err := s.authRepo.FindByUsername(ctx, user.Username)
		if err != nil {
			return nil, err
		}
		if !auth.IsActive || auth.IsDeleted {
			return nil, domain.ErrUserLocked
		}

		return s.generateLoginResponse(ctx, &auth, user, ip)
	}

	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		if !existingUser.IsActive || existingUser.IsDeleted {
			return nil, domain.ErrUserLocked
		}

		newOAuthAcc := &domain.OAuthAccount{
			UserID:         existingUser.ID,
			Provider:       "github",
			ProviderUserID: ghIDStr,
			Email:          email,
			AvatarURL:      strings.TrimSpace(ghUser.AvatarURL),
			AccessToken:    encryptedToken,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if err := s.oauthRepo.Create(ctx, newOAuthAcc); err != nil {
			return nil, err
		}

		auth, err := s.authRepo.FindByUsername(ctx, existingUser.Username)
		if err != nil {
			return nil, err
		}
		if !auth.IsActive || auth.IsDeleted {
			return nil, domain.ErrUserLocked
		}

		return s.generateLoginResponse(ctx, &auth, existingUser, ip)
	}

	role, err := s.getDefaultRole(ctx)
	if err != nil {
		return nil, err
	}
	if role == nil || !role.IsActive {
		return nil, domain.ErrRoleNotFound
	}

	uniqueUsername := s.generateUniqueUsername(ctx, ghUser.Login)

	fullName := strings.TrimSpace(ghUser.Name)
	if fullName == "" {
		fullName = uniqueUsername
	}
	if len(fullName) > 100 {
		fullName = fullName[:100]
	}

	dummyPassword := uuid.New().String() + "!Aa1"
	passwordHash, err := utils.HashPassword(dummyPassword)
	if err != nil {
		return nil, err
	}

	newUser := &domain.User{
		Username:  uniqueUsername,
		Email:     email,
		FullName:  fullName,
		AvatarURL: strings.TrimSpace(ghUser.AvatarURL),
		RoleID:    &role.ID,
		IsActive:  true,
		IsDeleted: false,
		CreatedBy: "oauth_github",
		UpdatedBy: "oauth_github",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := newUser.Validate(); err != nil {
		return nil, err
	}

	if err := s.userRepo.CreateWithAuth(ctx, newUser, passwordHash); err != nil {
		return nil, err
	}

	newOAuthAcc := &domain.OAuthAccount{
		UserID:         newUser.ID,
		Provider:       "github",
		ProviderUserID: ghIDStr,
		Email:          email,
		AvatarURL:      strings.TrimSpace(ghUser.AvatarURL),
		AccessToken:    encryptedToken,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := s.oauthRepo.Create(ctx, newOAuthAcc); err != nil {
		return nil, err
	}

	auth, err := s.authRepo.FindByUsername(ctx, newUser.Username)
	if err != nil {
		return nil, err
	}

	return s.generateLoginResponse(ctx, &auth, newUser, ip)
}

func (s *AuthService) generateUniqueUsername(ctx context.Context, baseLogin string) string {
	baseLogin = strings.TrimSpace(baseLogin)
	var validChars []rune
	for _, r := range baseLogin {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.' || r == '-' {
			validChars = append(validChars, r)
		}
	}
	cleaned := string(validChars)
	if len(cleaned) < 3 {
		cleaned = "user_" + cleaned
	}
	if len(cleaned) > 40 {
		cleaned = cleaned[:40]
	}

	candidate := cleaned
	existing, err := s.userRepo.FindByUsername(ctx, candidate)
	if err == nil && existing == nil {
		return candidate
	}

	for i := 1; i <= 100; i++ {
		candidate = fmt.Sprintf("%s_%d", cleaned, i)
		existing, err = s.userRepo.FindByUsername(ctx, candidate)
		if err == nil && existing == nil {
			return candidate
		}
	}

	return fmt.Sprintf("u_%d", time.Now().Unix())
}

func (s *AuthService) GetGitHubRepositories(ctx context.Context, userID int) ([]*domain.GitHubRepository, error) {
	if s.gitHubClient == nil || s.oauthRepo == nil {
		return nil, errors.New("github client or oauth repository is not initialized")
	}

	accounts, err := s.oauthRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var ghAcc *domain.OAuthAccount
	for _, acc := range accounts {
		if acc.Provider == "github" {
			ghAcc = &acc
			break
		}
	}

	if ghAcc == nil || ghAcc.AccessToken == "" {
		return nil, domain.ErrGitHubAccountNotLinked
	}

	secretKey := "oauth-secret-key"
	if s.config != nil && s.config.Server.Jwt.SecretKey != "" {
		secretKey = s.config.Server.Jwt.SecretKey
	}

	decryptedToken, err := utils.DecryptAESGCM(ghAcc.AccessToken, secretKey)
	if err != nil {
		return nil, domain.ErrGitHubTokenInvalid
	}

	return s.gitHubClient.GetUserRepositories(ctx, decryptedToken)
}

func (s *AuthService) CreateGitHubIssue(ctx context.Context, userID int, owner, repo string, req domain.GitHubCreateIssueRequest) (*domain.GitHubIssueResponse, error) {
	if s.gitHubClient == nil || s.oauthRepo == nil {
		return nil, errors.New("github client or oauth repository is not initialized")
	}

	accounts, err := s.oauthRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var ghAcc *domain.OAuthAccount
	for _, acc := range accounts {
		if acc.Provider == "github" {
			ghAcc = &acc
			break
		}
	}

	if ghAcc == nil || ghAcc.AccessToken == "" {
		return nil, domain.ErrGitHubAccountNotLinked
	}

	secretKey := "oauth-secret-key"
	if s.config != nil && s.config.Server.Jwt.SecretKey != "" {
		secretKey = s.config.Server.Jwt.SecretKey
	}

	decryptedToken, err := utils.DecryptAESGCM(ghAcc.AccessToken, secretKey)
	if err != nil {
		return nil, domain.ErrGitHubTokenInvalid
	}

	return s.gitHubClient.CreateIssue(ctx, decryptedToken, owner, repo, req)
}

