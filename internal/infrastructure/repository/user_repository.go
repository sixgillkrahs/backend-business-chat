package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/database"
)

type userRepository struct {
	db *database.PostgresDB
}

func NewUserRepository(db *database.PostgresDB) domain.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) CreateWithAuth(ctx context.Context, user *domain.User, passwordHash string) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	insertUserQuery := `
		INSERT INTO users (
			username, email, full_name, avatar_url, phone_number, address,
			role_id, created_by, updated_by, created_at, updated_at, is_active, is_deleted
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(ctx, insertUserQuery,
		user.Username, user.Email, user.FullName, user.AvatarURL, user.PhoneNumber, user.Address,
		user.RoleID, user.CreatedBy, user.UpdatedBy, user.CreatedAt, user.UpdatedAt, user.IsActive, user.IsDeleted,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "username") {
				return domain.ErrUsernameDuplicate
			}
			if strings.Contains(pgErr.ConstraintName, "email") {
				return domain.ErrEmailDuplicate
			}
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}

	authID := uuid.New().String()
	insertAuthQuery := `
		INSERT INTO auths (
			id, user_id, username, password_hash, role_id,
			created_by, updated_by, created_at, updated_at, is_active, is_deleted
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	roleID := 0
	if user.RoleID != nil {
		roleID = *user.RoleID
	}
	_, err = tx.Exec(ctx, insertAuthQuery,
		authID, user.ID, user.Username, passwordHash, roleID,
		user.CreatedBy, user.UpdatedBy, user.CreatedAt, user.UpdatedAt, user.IsActive, user.IsDeleted,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "username") {
				return domain.ErrUsernameDuplicate
			}
		}
		return fmt.Errorf("failed to insert auth: %w", err)
	}

	insertHistoryQuery := `
		INSERT INTO history_passwords (user_id, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = tx.Exec(ctx, insertHistoryQuery, user.ID, passwordHash, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert history password: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			username, email, full_name, avatar_url, phone_number, address,
			role_id, created_by, updated_by, created_at, updated_at, is_active, is_deleted
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`
	err := r.db.Pool.QueryRow(ctx, query,
		user.Username, user.Email, user.FullName, user.AvatarURL, user.PhoneNumber, user.Address,
		user.RoleID, user.CreatedBy, user.UpdatedBy, user.CreatedAt, user.UpdatedAt, user.IsActive, user.IsDeleted,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "username") {
				return domain.ErrUsernameDuplicate
			}
			if strings.Contains(pgErr.ConstraintName, "email") {
				return domain.ErrEmailDuplicate
			}
		}
		return err
	}
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id int) (*domain.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.full_name, COALESCE(u.avatar_url, ''), COALESCE(u.phone_number, ''),
		       COALESCE(u.address, ''), u.role_id, u.is_active, u.is_deleted, u.created_by, u.updated_by,
		       u.deleted_by, u.created_at, u.updated_at, u.deleted_at,
		       r.id, COALESCE(r.name, ''), COALESCE(r.description, ''), COALESCE(r.isactive, false), COALESCE(r.issystem, false),
		       r.created_at, r.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1 AND u.is_deleted = false
	`
	var u domain.User
	var rID *int
	var rName, rDesc string
	var rActive, rSystem bool
	var rCreatedAt, rUpdatedAt *time.Time

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.FullName, &u.AvatarURL, &u.PhoneNumber,
		&u.Address, &u.RoleID, &u.IsActive, &u.IsDeleted, &u.CreatedBy, &u.UpdatedBy,
		&u.DeletedBy, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
		&rID, &rName, &rDesc, &rActive, &rSystem, &rCreatedAt, &rUpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	if rID != nil {
		var roleCreatedAt, roleUpdatedAt time.Time
		if rCreatedAt != nil {
			roleCreatedAt = *rCreatedAt
		}
		if rUpdatedAt != nil {
			roleUpdatedAt = *rUpdatedAt
		}
		u.Role = &domain.Role{
			ID:          *rID,
			Name:        rName,
			Description: rDesc,
			IsActive:    rActive,
			IsSystem:    rSystem,
			CreatedAt:   roleCreatedAt,
			UpdatedAt:   roleUpdatedAt,
		}
	}

	return &u, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.full_name, COALESCE(u.avatar_url, ''), COALESCE(u.phone_number, ''),
		       COALESCE(u.address, ''), u.role_id, u.is_active, u.is_deleted, u.created_by, u.updated_by,
		       u.deleted_by, u.created_at, u.updated_at, u.deleted_at
		FROM users u
		WHERE u.username = $1 AND u.is_deleted = false
	`
	var u domain.User
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(
		&u.ID, &u.Username, &u.Email, &u.FullName, &u.AvatarURL, &u.PhoneNumber,
		&u.Address, &u.RoleID, &u.IsActive, &u.IsDeleted, &u.CreatedBy, &u.UpdatedBy,
		&u.DeletedBy, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}
	return &u, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.full_name, COALESCE(u.avatar_url, ''), COALESCE(u.phone_number, ''),
		       COALESCE(u.address, ''), u.role_id, u.is_active, u.is_deleted, u.created_by, u.updated_by,
		       u.deleted_by, u.created_at, u.updated_at, u.deleted_at
		FROM users u
		WHERE u.email = $1 AND u.is_deleted = false
	`
	var u domain.User
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.FullName, &u.AvatarURL, &u.PhoneNumber,
		&u.Address, &u.RoleID, &u.IsActive, &u.IsDeleted, &u.CreatedBy, &u.UpdatedBy,
		&u.DeletedBy, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return &u, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	updateUserQuery := `
		UPDATE users
		SET username = $1, email = $2, full_name = $3, avatar_url = $4, phone_number = $5, address = $6,
		    role_id = $7, is_active = $8, updated_by = $9, updated_at = $10
		WHERE id = $11 AND is_deleted = false
	`
	cmdTag, err := tx.Exec(ctx, updateUserQuery,
		user.Username, user.Email, user.FullName, user.AvatarURL, user.PhoneNumber, user.Address,
		user.RoleID, user.IsActive, user.UpdatedBy, user.UpdatedAt, user.ID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "username") {
				return domain.ErrUsernameDuplicate
			}
			if strings.Contains(pgErr.ConstraintName, "email") {
				return domain.ErrEmailDuplicate
			}
		}
		return fmt.Errorf("failed to update user: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	roleID := 0
	if user.RoleID != nil {
		roleID = *user.RoleID
	}
	updateAuthQuery := `
		UPDATE auths
		SET username = $1, role_id = $2, is_active = $3, updated_by = $4, updated_at = $5
		WHERE user_id = $6 AND is_deleted = false
	`
	_, err = tx.Exec(ctx, updateAuthQuery,
		user.Username, roleID, user.IsActive, user.UpdatedBy, user.UpdatedAt, user.ID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "username") {
				return domain.ErrUsernameDuplicate
			}
		}
		return fmt.Errorf("failed to sync auth credentials: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id int) error {
	query := `
		UPDATE users
		SET is_deleted = true, deleted_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *userRepository) Count(ctx context.Context, search string) (int64, error) {
	var query string
	var args []interface{}

	if search != "" {
		query = `
			SELECT COUNT(*) FROM users
			WHERE is_deleted = false AND (username ILIKE $1 OR email ILIKE $1 OR full_name ILIKE $1)
		`
		args = []interface{}{"%" + search + "%"}
	} else {
		query = `SELECT COUNT(*) FROM users WHERE is_deleted = false`
	}

	var count int64
	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *userRepository) GetUsersPaging(ctx context.Context, search string, offset, limit int) ([]domain.User, error) {
	var query string
	var args []interface{}

	baseSelect := `
		SELECT u.id, u.username, u.email, u.full_name, COALESCE(u.avatar_url, ''), COALESCE(u.phone_number, ''),
		       COALESCE(u.address, ''), u.role_id, u.is_active, u.is_deleted, u.created_by, u.updated_by,
		       u.deleted_by, u.created_at, u.updated_at, u.deleted_at,
		       r.id, COALESCE(r.name, ''), COALESCE(r.description, ''), COALESCE(r.isactive, false), COALESCE(r.issystem, false),
		       r.created_at, r.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.is_deleted = false
	`
	if search != "" {
		query = baseSelect + `
			AND (u.username ILIKE $1 OR u.email ILIKE $1 OR u.full_name ILIKE $1)
			ORDER BY u.id ASC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{"%" + search + "%", limit, offset}
	} else {
		query = baseSelect + `
			ORDER BY u.id ASC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		var rID *int
		var rName, rDesc string
		var rActive, rSystem bool
		var rCreatedAt, rUpdatedAt *time.Time

		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.FullName, &u.AvatarURL, &u.PhoneNumber,
			&u.Address, &u.RoleID, &u.IsActive, &u.IsDeleted, &u.CreatedBy, &u.UpdatedBy,
			&u.DeletedBy, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
			&rID, &rName, &rDesc, &rActive, &rSystem, &rCreatedAt, &rUpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if rID != nil {
			var roleCreatedAt, roleUpdatedAt time.Time
			if rCreatedAt != nil {
				roleCreatedAt = *rCreatedAt
			}
			if rUpdatedAt != nil {
				roleUpdatedAt = *rUpdatedAt
			}
			u.Role = &domain.Role{
				ID:          *rID,
				Name:        rName,
				Description: rDesc,
				IsActive:    rActive,
				IsSystem:    rSystem,
				CreatedAt:   roleCreatedAt,
				UpdatedAt:   roleUpdatedAt,
			}
		}

		users = append(users, u)
	}
	return users, rows.Err()
}
