package repository

import (
	"context"
	"errors"
	"fmt"
	"shorter-url/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type passwordResetTokensRepository struct {
	db *pgxpool.Pool
}

func NewPasswordResetTokensRepository(db *pgxpool.Pool) domain.PasswordResetTokensRepository {
	return &passwordResetTokensRepository{
		db: db,
	}
}

func (r *passwordResetTokensRepository) Create(ctx context.Context, passwordResetToken *domain.PasswordResetTokens) (*domain.PasswordResetTokens, error) {
	query := `INSERT INTO password_reset_tokens (user_id, token, expired_at) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRow(ctx, query, passwordResetToken.UserId, passwordResetToken.Token, passwordResetToken.ExpiredAt).Scan(&passwordResetToken.Id)
	if err != nil {
		return nil, fmt.Errorf("something wrong when create password reset token : %w", err)
	}
	return passwordResetToken, nil
}

func (r *passwordResetTokensRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM password_reset_tokens WHERE id = $1"

	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("something wrong when delete password reset token : %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("there is no data deleted, password reset token with ID %d not found", id)
	}

	return nil
}

func (r *passwordResetTokensRepository) FindByToken(ctx context.Context, token string) (*domain.PasswordResetTokens, error) {
	query := "SELECT id, user_id, token, expired_at FROM password_reset_tokens WHERE token = $1"
	var passwordResetToken domain.PasswordResetTokens
	err := r.db.QueryRow(ctx, query, token).Scan(&passwordResetToken.Id, &passwordResetToken.UserId, &passwordResetToken.Token, &passwordResetToken.ExpiredAt)
	if err != nil {
		// Jika errornya adalah karena datanya memang tidak ada
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("token not found")
		}
		// Jika error karena masalah teknis database (koneksi terputus, dll)
		return nil, fmt.Errorf("something wrong when find password reset token by token : %w", err)
	}
	return &passwordResetToken, nil
}
