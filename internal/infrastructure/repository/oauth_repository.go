package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/database"
)

type oauthRepository struct {
	db *database.PostgresDB
}

func NewOAuthRepository(db *database.PostgresDB) domain.OAuthRepository {
	return &oauthRepository{db: db}
}

func (r *oauthRepository) FindByProviderAndUserID(ctx context.Context, provider, providerUserID string) (*domain.OAuthAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, COALESCE(email, ''), COALESCE(avatar_url, ''), COALESCE(access_token, ''), created_at, updated_at
		FROM oauth_accounts
		WHERE provider = $1 AND provider_user_id = $2
	`
	var acc domain.OAuthAccount
	err := r.db.Pool.QueryRow(ctx, query, provider, providerUserID).Scan(
		&acc.ID, &acc.UserID, &acc.Provider, &acc.ProviderUserID, &acc.Email, &acc.AvatarURL, &acc.AccessToken, &acc.CreatedAt, &acc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find oauth account: %w", err)
	}
	return &acc, nil
}

func (r *oauthRepository) Create(ctx context.Context, account *domain.OAuthAccount) error {
	query := `
		INSERT INTO oauth_accounts (user_id, provider, provider_user_id, email, avatar_url, access_token, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	err := r.db.Pool.QueryRow(ctx, query,
		account.UserID, account.Provider, account.ProviderUserID, account.Email, account.AvatarURL, account.AccessToken,
		account.CreatedAt, account.UpdatedAt,
	).Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert oauth account: %w", err)
	}
	return nil
}

func (r *oauthRepository) UpdateAccessToken(ctx context.Context, id int, encryptedToken string) error {
	query := `
		UPDATE oauth_accounts
		SET access_token = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	_, err := r.db.Pool.Exec(ctx, query, encryptedToken, id)
	if err != nil {
		return fmt.Errorf("failed to update oauth access token: %w", err)
	}
	return nil
}

func (r *oauthRepository) FindByUserID(ctx context.Context, userID int) ([]domain.OAuthAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, COALESCE(email, ''), COALESCE(avatar_url, ''), COALESCE(access_token, ''), created_at, updated_at
		FROM oauth_accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query oauth accounts by user id: %w", err)
	}
	defer rows.Close()

	var accounts []domain.OAuthAccount
	for rows.Next() {
		var acc domain.OAuthAccount
		if err := rows.Scan(
			&acc.ID, &acc.UserID, &acc.Provider, &acc.ProviderUserID, &acc.Email, &acc.AvatarURL, &acc.AccessToken, &acc.CreatedAt, &acc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan oauth account: %w", err)
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}
