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


// Count counts the total number of permissions matching the search filter (OWASP A03 Parameterized query).
func (r *permissionRepository) Count(ctx context.Context, search string) (int64, error) {
	var query string
	var args []interface{}

	if search != "" {
		query = `SELECT COUNT(*) FROM permission WHERE name ILIKE $1 OR COALESCE(description, '') ILIKE $1`
		args = []interface{}{"%" + search + "%"}
	} else {
		query = `SELECT COUNT(*) FROM permission`
	}

	var count int64
	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count permissions: %w", err)
	}
	return count, nil
}

// GetPermissionsPaging retrieves paginated permissions and batch-loads their policies in 2 optimized queries (avoids N+1 query and prevents SQL Cartesian limits).
func (r *permissionRepository) GetPermissionsPaging(ctx context.Context, search string, offset, limit int) ([]domain.Permission, error) {
	var query string
	var args []interface{}

	baseSelect := `
		SELECT id, name, COALESCE(description, ''), created_at, updated_at
		FROM permission
	`
	if search != "" {
		query = baseSelect + `
			WHERE name ILIKE $1 OR COALESCE(description, '') ILIKE $1
			ORDER BY id ASC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{"%" + search + "%", limit, offset}
	} else {
		query = baseSelect + `
			ORDER BY id ASC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	permissions := make([]domain.Permission, 0)
	permIDs := make([]int, 0)
	permMap := make(map[int]int)

	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		p.Policies = make([]domain.Policy, 0)
		permMap[p.ID] = len(permissions)
		permissions = append(permissions, p)
		permIDs = append(permIDs, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(permIDs) == 0 {
		return permissions, nil
	}

	// Batch load all policies for the retrieved permissions in a single query
	policiesQuery := `
		SELECT 
			pp.permission_id,
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
		WHERE pp.permission_id = ANY($1)
		ORDER BY pp.permission_id ASC, p.id ASC
	`
	pRows, err := r.db.Pool.Query(ctx, policiesQuery, permIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies for permissions batch: %w", err)
	}
	defer pRows.Close()

	for pRows.Next() {
		var permID int
		var pol domain.Policy
		if err := pRows.Scan(
			&permID,
			&pol.ID,
			&pol.Name,
			&pol.Description,
			&pol.ActionID,
			&pol.ResourceID,
			&pol.ActionName,
			&pol.ResourceName,
			&pol.CreatedAt,
			&pol.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan policy in permission batch: %w", err)
		}

		if idx, ok := permMap[permID]; ok {
			permissions[idx].Policies = append(permissions[idx].Policies, pol)
		}
	}

	if err := pRows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}
