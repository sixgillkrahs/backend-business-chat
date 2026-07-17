package application

import (
	"context"

	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
)

type AuthService struct {
	authRepo      domain.ActionRepository
	resourcesRepo domain.ResourceRepository
	policyRepo    domain.PolicyRepository
}

func NewAuthService(authRepo domain.ActionRepository, resourcesRepo domain.ResourceRepository, policyRepo domain.PolicyRepository) *AuthService {
	return &AuthService{authRepo: authRepo, resourcesRepo: resourcesRepo, policyRepo: policyRepo}
}

func (s *AuthService) GetAllActions(ctx context.Context) ([]domain.Action, error) {
	return s.authRepo.GetAllActions(ctx)
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

func (s *AuthService) GetPolicies(ctx context.Context, page, limit int) ([]domain.Policy, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	return s.policyRepo.GetPoliciesPage(ctx, offset, limit)
}

func (s *AuthService) CountPolicies(ctx context.Context) (int64, error) {
	return s.policyRepo.Count(ctx)
}
