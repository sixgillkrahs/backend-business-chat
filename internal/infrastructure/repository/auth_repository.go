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

type resourceRepository struct {
	baseRepository[domain.Resource]
}

func NewResourceRepository(db *database.PostgresDB) domain.ResourceRepository {
	return &resourceRepository{
		baseRepository: newBaseRepository[domain.Resource](db),
	}
}

func (r *resourceRepository) GetAllResources(ctx context.Context) ([]domain.Resource, error) {
	return r.baseRepository.GetAll(ctx, domain.Resource{})
}

func (r *resourceRepository) Upsert(ctx context.Context, resource *domain.Resource) error {
	query := `
		INSERT INTO resources (name, code, description, created_at, updated_at)
		VALUES ($1, $2, $3, COALESCE($4, NOW()), COALESCE($5, NOW()))
		ON CONFLICT (code) DO UPDATE
		SET name = EXCLUDED.name,
		    description = EXCLUDED.description,
		    updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	var createdAt, updatedAt interface{}
	if resource.CreatedAt.IsZero() {
		createdAt = nil
	} else {
		createdAt = resource.CreatedAt
	}
	if resource.UpdatedAt.IsZero() {
		updatedAt = nil
	} else {
		updatedAt = resource.UpdatedAt
	}

	err := r.baseRepository.db.Pool.QueryRow(ctx, query, resource.Name, resource.Code, resource.Description, createdAt, updatedAt).Scan(&resource.ID, &resource.CreatedAt, &resource.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

type policyRepository struct {
	baseRepository[domain.Policy]
}

func NewPolicyRepository(db *database.PostgresDB) domain.PolicyRepository {
	return &policyRepository{
		baseRepository: newBaseRepository[domain.Policy](db),
	}
}

func (r *policyRepository) GetPoliciesPaging(ctx context.Context, offset, limit int) ([]domain.Policy, error) {
	return r.baseRepository.GetPage(ctx, domain.Policy{}, offset, limit)
}

func (r *policyRepository) Count(ctx context.Context) (int64, error) {
	return r.baseRepository.Count(ctx)
}
