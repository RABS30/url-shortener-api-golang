package repository

import (
	"context"
	"errors"
	"fmt"
	"shorter-url/internal/database"
	"shorter-url/internal/domain"

	"github.com/jackc/pgx/v5"
)

type oauthAccountsRepository struct {
	db database.PgxDatabase
}

func NewOauthAccountsRepository(db database.PgxDatabase) domain.OauthAccountsRepository {
	return &oauthAccountsRepository{
		db: db,
	}
}

func (r *oauthAccountsRepository) Upsert(ctx context.Context, oauthAccount *domain.OauthAccounts) (*domain.OauthAccounts, error) {
	query := `INSERT INTO oauth_accounts (user_id, provider, provider_user_id) VALUES ($1, $2, $3) ON CONFLICT (provider, provider_user_id) DO UPDATE SET updated_at = NOW() RETURNING id, user_id, provider, provider_user_id, created_at, updated_at`

	var result domain.OauthAccounts

	err := r.db.QueryRow(ctx, query, oauthAccount.UserId, oauthAccount.Provider, oauthAccount.ProviderUserId).Scan(
		&result.Id,
		&result.UserId,
		&result.Provider,
		&result.ProviderUserId,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert oauth account: %w", err)
	}

	return &result, nil
}

func (r *oauthAccountsRepository) FindById(ctx context.Context, id int64) (*domain.OauthAccounts, error) {
	query := `SELECT id, user_id, provider, provider_user_id, created_at, updated_at FROM oauth_accounts WHERE id = $1`

	var result domain.OauthAccounts

	err := r.db.QueryRow(ctx, query, id).Scan(
		&result.Id,
		&result.UserId,
		&result.Provider,
		&result.ProviderUserId,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find oauth account by id: %w", err)
	}

	return &result, nil
}

func (r *oauthAccountsRepository) FindByProviderUserId(ctx context.Context, provider, providerUserId string) (*domain.OauthAccounts, error) {
	query := `SELECT id, user_id, provider, provider_user_id, created_at, updated_at FROM oauth_accounts WHERE provider = $1 AND provider_user_id = $2`

	var result domain.OauthAccounts

	err := r.db.QueryRow(ctx, query, provider, providerUserId).Scan(
		&result.Id,
		&result.UserId,
		&result.Provider,
		&result.ProviderUserId,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find oauth account by provider and provider user id: %w", err)
	}

	return &result, nil
}

func (r *oauthAccountsRepository) FindByUserIdAndProvider(ctx context.Context, userId int64, provider string) (*domain.OauthAccounts, error) {
	query := `SELECT id, user_id, provider, provider_user_id, created_at, updated_at FROM oauth_accounts WHERE user_id = $1 AND provider = $2`

	var result domain.OauthAccounts

	err := r.db.QueryRow(ctx, query, userId, provider).Scan(
		&result.Id,
		&result.UserId,
		&result.Provider,
		&result.ProviderUserId,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find oauth account by user id and provider: %w", err)
	}

	return &result, nil
}

func (r *oauthAccountsRepository) Update(ctx context.Context, oauthAccount *domain.OauthAccounts) (*domain.OauthAccounts, error) {
	query := `UPDATE oauth_accounts SET provider = $1, provider_user_id = $2, updated_at = NOW() WHERE id = $3 RETURNING id, user_id, provider, provider_user_id, created_at, updated_at`

	var result domain.OauthAccounts

	err := r.db.QueryRow(ctx, query, oauthAccount.Provider, oauthAccount.ProviderUserId, oauthAccount.Id).Scan(
		&result.Id,
		&result.UserId,
		&result.Provider,
		&result.ProviderUserId,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("update oauth account: row with id %d not found", oauthAccount.Id)
		}
		return nil, fmt.Errorf("update oauth account: %w", err)
	}

	return &result, nil
}

func (r *oauthAccountsRepository) DeleteById(ctx context.Context, id int64) error {
	query := `DELETE FROM oauth_accounts WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete oauth account by id: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("delete oauth account by id: row with id %d not found", id)
	}

	return nil
}
