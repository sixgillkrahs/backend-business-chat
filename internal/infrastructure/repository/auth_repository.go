package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/database"
)

type actionRepository struct {
	baseRepository[domain.Action]
}

func NewActionRepository(db *database.PostgresDB) domain.ActionRepository {
	return &actionRepository{
		baseRepository: newBaseRepository[domain.Action](db),
	}
}

func (r *actionRepository) GetAllActions(ctx context.Context) ([]domain.Action, error) {
	return r.baseRepository.GetAll(ctx, domain.Action{})
}

func (r *actionRepository) GetActionById(ctx context.Context, id int) (domain.Action, error) {
	action, err := r.baseRepository.FindByID(ctx, int64(id))
	if err != nil {
		return domain.Action{}, nil
	}
	return *action, nil
}

type resourceRepository struct {
	baseRepository[domain.Resource]
}

func NewResourceRepository(db *database.PostgresDB) domain.ResourceRepository {
	return &resourceRepository{
		baseRepository: newBaseRepository[domain.Resource](db),
	}
}

func (r *resourceRepository) GetResourceById(ctx context.Context, id int) (domain.Resource, error) {
	action, err := r.baseRepository.FindByID(ctx, int64(id))
	if err != nil {
		return domain.Resource{}, nil
	}
	return *action, nil
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

func (r *policyRepository) GetPoliciesPaging(ctx context.Context, search string, offset, limit int) ([]domain.Policy, error) {
	var query string
	var args []interface{}

	baseSelect := `
		SELECT 
			p.id, 
			p.name, 
			COALESCE(p.description, ''), 
			p.action_id, 
			p.resource_id, 
			COALESCE(a.name, '') as action_name, 
			COALESCE(r.name, '') as resource_name, 
			p.created_at, 
			p.updated_at
		FROM policies p
		LEFT JOIN actions a ON p.action_id = a.id
		LEFT JOIN resources r ON p.resource_id = r.id
	`

	if search != "" {
		query = baseSelect + `
			WHERE p.name ILIKE $1 OR COALESCE(p.description, '') ILIKE $1 OR a.name ILIKE $1 OR r.name ILIKE $1
			ORDER BY p.id ASC
			LIMIT $2 OFFSET $3
		`
		pattern := "%" + search + "%"
		args = []interface{}{pattern, limit, offset}
	} else {
		query = baseSelect + `
			ORDER BY p.id ASC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	policies := make([]domain.Policy, 0)
	for rows.Next() {
		var policy domain.Policy
		err := rows.Scan(
			&policy.ID,
			&policy.Name,
			&policy.Description,
			&policy.ActionID,
			&policy.ResourceID,
			&policy.ActionName,
			&policy.ResourceName,
			&policy.CreatedAt,
			&policy.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return policies, nil
}

func (r *policyRepository) Count(ctx context.Context, search string) (int64, error) {
	var query string
	var args []interface{}

	if search != "" {
		query = `
			SELECT COUNT(*)
			FROM policies p
			LEFT JOIN actions a ON p.action_id = a.id
			LEFT JOIN resources r ON p.resource_id = r.id
			WHERE p.name ILIKE $1 OR COALESCE(p.description, '') ILIKE $1 OR a.name ILIKE $1 OR r.name ILIKE $1
		`
		args = []interface{}{"%" + search + "%"}
	} else {
		query = `SELECT COUNT(*) FROM policies`
	}

	var count int64
	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *policyRepository) Insert(ctx context.Context, policy *domain.Policy) error {
	var err = r.baseRepository.Create(ctx, policy)
	if err != nil {
		return err
	}
	return nil
}

func (r *policyRepository) GetPolicyById(ctx context.Context, id int) (domain.Policy, error) {
	query := `
		SELECT 
			p.id, 
			p.name, 
			COALESCE(p.description, ''), 
			p.action_id, 
			p.resource_id, 
			COALESCE(a.name, '') as action_name, 
			COALESCE(r.name, '') as resource_name, 
			p.created_at, 
			p.updated_at
		FROM policies p
		LEFT JOIN resources r ON p.resource_id = r.id
		LEFT JOIN actions a ON p.action_id = a.id
		WHERE p.id = $1
	`
	var policy domain.Policy
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&policy.ID,
		&policy.Name,
		&policy.Description,
		&policy.ActionID,
		&policy.ResourceID,
		&policy.ActionName,
		&policy.ResourceName,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)
	if err != nil {
		return domain.Policy{}, err
	}

	return policy, nil
}

type authRepository struct {
	baseRepository[domain.Auth]
}

func NewAuthRepository(db *database.PostgresDB) domain.AuthRepository {
	return &authRepository{
		baseRepository: newBaseRepository[domain.Auth](db),
	}
}

func (r *authRepository) FindByUsername(ctx context.Context, username string) (domain.Auth, error) {
	authPtr, err := r.baseRepository.FindOneByField(ctx, "username", username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Auth{}, domain.ErrUserNotFound
		}
		return domain.Auth{}, err
	}
	if authPtr == nil {
		return domain.Auth{}, domain.ErrUserNotFound
	}

	return *authPtr, nil
}
