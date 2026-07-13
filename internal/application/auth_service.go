package application

import (
	"context"

	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
)

type AuthService struct {
	authRepo domain.ActionRepository
}

func NewAuthService(authRepo domain.ActionRepository) *AuthService {
	return &AuthService{authRepo: authRepo}
}

func (s *AuthService) GetAllActions(ctx context.Context) ([]domain.Action, error) {
	return s.authRepo.GetAllActions(ctx)
}
