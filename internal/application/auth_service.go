package application

import (
	"context"

	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
)

type AuthService struct {
	authRepo      domain.ActionRepository
	resourcesRepo domain.ResourceRepository
}

func NewAuthService(authRepo domain.ActionRepository, resourcesRepo domain.ResourceRepository) *AuthService {
	return &AuthService{authRepo: authRepo, resourcesRepo: resourcesRepo}
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
