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
}

func NewAuthService(cfg *config.Config, actionRepo domain.ActionRepository, authRepo domain.AuthRepository, resourcesRepo domain.ResourceRepository, policyRepo domain.PolicyRepository, permissionRepo domain.PermissionRepository) *AuthService {
	return &AuthService{
		config:         cfg,
		actionRepo:     actionRepo,
		authRepo:       authRepo,
		resourcesRepo:  resourcesRepo,
		policyRepo:     policyRepo,
		permissionRepo: permissionRepo,
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
	// OWASP A04: Input boundary sanitization and defense in depth
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50 // Enforce max page limit to prevent memory exhaustion / DoS
	}
	if offset < 0 {
		offset = 0
	}
	search = strings.TrimSpace(search)
	if len(search) > 100 {
		search = search[:100] // Cap search string length
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
