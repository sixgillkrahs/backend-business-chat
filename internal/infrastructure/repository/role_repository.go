package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/database"
)

type roleRepository struct {
	db *database.PostgresDB
}

func NewRoleRepository(db *database.PostgresDB) domain.RoleRepository {
	return &roleRepository{
		db: db,
	}
}

func (r *roleRepository) ValidatePermissionIDs(ctx context.Context, permissionIDs []int) (bool, error) {
	if len(permissionIDs) == 0 {
		return false, nil
	}
	query := `SELECT COUNT(DISTINCT id) FROM permission WHERE id = ANY($1)`
	var count int
	err := r.db.Pool.QueryRow(ctx, query, permissionIDs).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to validate permission ids: %w", err)
	}
	return count == len(permissionIDs), nil
}

func (r *roleRepository) CreateWithPermissions(ctx context.Context, role *domain.Role, permissionIDs []int) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	insertRoleQuery := `
		INSERT INTO roles (name, description, isactive, issystem, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(ctx, insertRoleQuery, role.Name, role.Description, role.IsActive, role.IsSystem, role.CreatedAt, role.UpdatedAt).
		Scan(&role.ID, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert role: %w", err)
	}

	if len(permissionIDs) > 0 {
		insertBatchQuery := `
			INSERT INTO role_permissions (role_id, permission_id)
			SELECT $1, UNNEST($2::int[])
		`
		if _, err := tx.Exec(ctx, insertBatchQuery, role.ID, permissionIDs); err != nil {
			return fmt.Errorf("failed to batch link permissions to role: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit role creation transaction: %w", err)
	}
	return nil
}

func (r *roleRepository) UpdateWithPermissions(ctx context.Context, role *domain.Role, permissionIDs []int) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	updateRoleQuery := `
		UPDATE roles
		SET name = $1, description = $2, isactive = $3, updated_at = $4
		WHERE id = $5
	`
	tag, err := tx.Exec(ctx, updateRoleQuery, role.Name, role.Description, role.IsActive, role.UpdatedAt, role.ID)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrRoleNotFound
	}

	if permissionIDs != nil {
		deleteLinkQuery := `DELETE FROM role_permissions WHERE role_id = $1`
		if _, err := tx.Exec(ctx, deleteLinkQuery, role.ID); err != nil {
			return fmt.Errorf("failed to delete old role permissions: %w", err)
		}

		if len(permissionIDs) > 0 {
			insertBatchQuery := `
				INSERT INTO role_permissions (role_id, permission_id)
				SELECT $1, UNNEST($2::int[])
			`
			if _, err := tx.Exec(ctx, insertBatchQuery, role.ID, permissionIDs); err != nil {
				return fmt.Errorf("failed to batch link permissions to role: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit role update transaction: %w", err)
	}
	return nil
}

func (r *roleRepository) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	query := `SELECT id, name, COALESCE(description, ''), isactive, issystem, created_at, updated_at FROM roles WHERE name = $1`
	var role domain.Role
	err := r.db.Pool.QueryRow(ctx, query, name).Scan(&role.ID, &role.Name, &role.Description, &role.IsActive, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query role by name: %w", err)
	}
	return &role, nil
}

func (r *roleRepository) GetRoleById(ctx context.Context, id int) (*domain.Role, error) {
	queryRole := `SELECT id, name, COALESCE(description, ''), isactive, issystem, created_at, updated_at FROM roles WHERE id = $1`
	var role domain.Role
	err := r.db.Pool.QueryRow(ctx, queryRole, id).Scan(&role.ID, &role.Name, &role.Description, &role.IsActive, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query role by id: %w", err)
	}

	queryPerms := `
		SELECT 
			p.id, 
			p.name, 
			COALESCE(p.description, ''), 
			p.created_at, 
			p.updated_at
		FROM role_permissions rp
		JOIN permission p ON rp.permission_id = p.id
		WHERE rp.role_id = $1
		ORDER BY p.id ASC
	`
	rows, err := r.db.Pool.Query(ctx, queryPerms, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions for role: %w", err)
	}
	defer rows.Close()

	perms := make([]domain.Permission, 0)
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan permission in role: %w", err)
		}
		perms = append(perms, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	role.Permissions = perms
	return &role, nil
}

func (r *roleRepository) Count(ctx context.Context, search string) (int64, error) {
	var query string
	var args []interface{}

	if search != "" {
		query = `SELECT COUNT(*) FROM roles WHERE name ILIKE $1 OR COALESCE(description, '') ILIKE $1`
		args = []interface{}{"%" + search + "%"}
	} else {
		query = `SELECT COUNT(*) FROM roles`
	}

	var count int64
	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count roles: %w", err)
	}
	return count, nil
}

func (r *roleRepository) GetRolesPaging(ctx context.Context, search string, offset, limit int) ([]domain.Role, error) {
	var query string
	var args []interface{}

	baseSelect := `
		SELECT id, name, COALESCE(description, ''), isactive, issystem, created_at, updated_at
		FROM roles
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
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	roles := make([]domain.Role, 0)
	roleIDs := make([]int, 0)
	roleMap := make(map[int]int)

	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.IsActive,
			&role.IsSystem,
			&role.CreatedAt,
			&role.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		role.Permissions = make([]domain.Permission, 0)
		roleMap[role.ID] = len(roles)
		roles = append(roles, role)
		roleIDs = append(roleIDs, role.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(roleIDs) == 0 {
		return roles, nil
	}

	permsQuery := `
		SELECT 
			rp.role_id,
			p.id, 
			p.name, 
			COALESCE(p.description, ''), 
			p.created_at, 
			p.updated_at
		FROM role_permissions rp
		JOIN permission p ON rp.permission_id = p.id
		WHERE rp.role_id = ANY($1)
		ORDER BY rp.role_id ASC, p.id ASC
	`
	pRows, err := r.db.Pool.Query(ctx, permsQuery, roleIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions for roles batch: %w", err)
	}
	defer pRows.Close()

	for pRows.Next() {
		var roleID int
		var perm domain.Permission
		if err := pRows.Scan(
			&roleID,
			&perm.ID,
			&perm.Name,
			&perm.Description,
			&perm.CreatedAt,
			&perm.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan permission in role batch: %w", err)
		}

		if idx, ok := roleMap[roleID]; ok {
			roles[idx].Permissions = append(roles[idx].Permissions, perm)
		}
	}
	if err := pRows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}
