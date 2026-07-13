package repository

import (
	"context"

	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/database"
)

type authRepository struct {
	baseRepository[domain.Action]
}

func NewAuthRepository(db *database.PostgresDB) domain.ActionRepository {
	return &authRepository{
		baseRepository: newBaseRepository[domain.Action](db),
	}
}

func (r *authRepository) GetAllActions(ctx context.Context) ([]domain.Action, error) {
	return r.baseRepository.GetAll(ctx, domain.Action{})
}
