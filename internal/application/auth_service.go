package application

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/sixgillkrahs/backend-business-chat/internal/application/dto"
	"github.com/sixgillkrahs/backend-business-chat/internal/config"
	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
	"github.com/sixgillkrahs/backend-business-chat/pkg/utils"
)

type AuthService struct {
	config        *config.Config
	actionRepo    domain.ActionRepository
	authRepo      domain.AuthRepository
	resourcesRepo domain.ResourceRepository
	policyRepo    domain.PolicyRepository
}

func NewAuthService(cfg *config.Config, actionRepo domain.ActionRepository, authRepo domain.AuthRepository, resourcesRepo domain.ResourceRepository, policyRepo domain.PolicyRepository) *AuthService {
	return &AuthService{
		config:        cfg,
		actionRepo:    actionRepo,
		authRepo:      authRepo,
		resourcesRepo: resourcesRepo,
		policyRepo:    policyRepo,
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

func (s *AuthService) GetPoliciesPaging(ctx context.Context, page, limit int) ([]domain.Policy, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	return s.policyRepo.GetPoliciesPaging(ctx, offset, limit)
}

func (s *AuthService) CountPolicies(ctx context.Context) (int64, error) {
	return s.policyRepo.Count(ctx)
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
