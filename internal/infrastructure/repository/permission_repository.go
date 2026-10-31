package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/database"
)

type permissionRepository struct {
	db *database.PostgresDB
}

func NewPermissionRepository(db *database.PostgresDB) domain.PermissionRepository {
	return &permissionRepository{
		db: db,
	}
}

func (r *permissionRepository) ValidatePolicyIDs(ctx context.Context, policyIDs []int) (bool, error) {
	if len(policyIDs) == 0 {
		return false, nil
	}
	query := `SELECT COUNT(DISTINCT id) FROM policies WHERE id = ANY($1)`
	var count int
	err := r.db.Pool.QueryRow(ctx, query, policyIDs).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to validate policy ids: %w", err)
	}
	return count == len(policyIDs), nil
}

func (r *permissionRepository) CreateWithPolicies(ctx context.Context, permission *domain.Permission, policyIDs []int) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	insertPermQuery := `
		INSERT INTO permission (name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(ctx, insertPermQuery, permission.Name, permission.Description, permission.CreatedAt, permission.UpdatedAt).
		Scan(&permission.ID, &permission.CreatedAt, &permission.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert permission: %w", err)
	}

	insertLinkQuery := `INSERT INTO permission_policies (permission_id, policy_id) VALUES ($1, $2)`
	for _, policyID := range policyIDs {
		if _, err := tx.Exec(ctx, insertLinkQuery, permission.ID, policyID); err != nil {
			return fmt.Errorf("failed to link policy %d to permission: %w", policyID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit permission creation transaction: %w", err)
	}
	return nil
}

func (r *permissionRepository) GetPermissionByName(ctx context.Context, name string) (*domain.Permission, error) {
	query := `SELECT id, name, COALESCE(description, ''), created_at, updated_at FROM permission WHERE name = $1`
	var p domain.Permission
	err := r.db.Pool.QueryRow(ctx, query, name).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query permission by name: %w", err)
	}
	return &p, nil
}

func (r *permissionRepository) GetPermissionById(ctx context.Context, id int) (*domain.Permission, error) {
	queryPerm := `SELECT id, name, COALESCE(description, ''), created_at, updated_at FROM permission WHERE id = $1`
	var p domain.Permission
	err := r.db.Pool.QueryRow(ctx, queryPerm, id).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query permission by id: %w", err)
	}

	queryPolicies := `
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
		FROM permission_policies pp
		JOIN policies p ON pp.policy_id = p.id
		LEFT JOIN actions a ON p.action_id = a.id
		LEFT JOIN resources r ON p.resource_id = r.id
		WHERE pp.permission_id = $1
		ORDER BY p.id ASC
	`
	rows, err := r.db.Pool.Query(ctx, queryPolicies, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies for permission: %w", err)
	}
	defer rows.Close()

	policies := make([]domain.Policy, 0)
	for rows.Next() {
		var policy domain.Policy
		if err := rows.Scan(
			&policy.ID,
			&policy.Name,
			&policy.Description,
			&policy.ActionID,
			&policy.ResourceID,
			&policy.ActionName,
			&policy.ResourceName,
			&policy.CreatedAt,
			&policy.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}
		policies = append(policies, policy)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	p.Policies = policies
	return &p, nil
}
