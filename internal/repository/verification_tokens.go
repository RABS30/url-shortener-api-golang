package repository

import (
	"context"
	"errors"
	"fmt"
	"shorter-url/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type verificationTokenRepository struct {
	db *pgxpool.Pool
}

func NewVerificationTokenRepository(db *pgxpool.Pool) domain.VerificationTokenRepository {
	return &verificationTokenRepository{
		db: db,
	}
}

func (v *verificationTokenRepository) Create(ctx context.Context, verificationToken *domain.VerificationToken) (*domain.VerificationToken, error) {
	err := v.db.QueryRow(ctx, "INSERT INTO verification_tokens (user_id, token, expired_at) VALUES ($1, $2, $3) RETURNING id, user_id, token, expired_at", verificationToken.UserId, verificationToken.Token, verificationToken.ExpiredAt).Scan(&verificationToken.Id, &verificationToken.UserId, &verificationToken.Token, &verificationToken.ExpiredAt)
	if err != nil {
		return nil, fmt.Errorf("something wrong when create verification token: %w", err)
	}

	return verificationToken, nil
}

func (v *verificationTokenRepository) Delete(ctx context.Context, id int64) error {
	commandTag, err := v.db.Exec(ctx, "DELETE FROM verification_tokens WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("something wrong when delete verification token: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("verification token with id %d not found", id)
	}

	return nil
}

func (v *verificationTokenRepository) FindByToken(ctx context.Context, token string) (*domain.VerificationToken, error) {
	var verificationToken domain.VerificationToken
	err := v.db.QueryRow(ctx, "SELECT id, user_id, token, expired_at FROM verification_tokens WHERE token = $1", token).Scan(&verificationToken.Id, &verificationToken.UserId, &verificationToken.Token, &verificationToken.ExpiredAt)
	if err != nil {
		// Mengisolasi logika bisnis jika token fiktif/tidak terdaftar
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("verification token not found")
		}
		return nil, fmt.Errorf("something wrong when find verification token by token : %w", err)
	}

	return &verificationToken, nil
}
