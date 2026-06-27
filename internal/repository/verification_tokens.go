package repository

import (
	"context"
	"errors"
	"fmt"
	"shorter-url/internal/database"
	"shorter-url/internal/domain"

	"github.com/jackc/pgx/v5"
)

type verificationTokenRepository struct {
	db database.PgxDatabase
}

func NewVerificationTokenRepository(db database.PgxDatabase) domain.VerificationTokenRepository {
	return &verificationTokenRepository{
		db: db,
	}
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

func (v *verificationTokenRepository) Create(ctx context.Context, verificationToken *domain.VerificationToken) (*domain.VerificationToken, error) {
	// Tambahkan created_at ke dalam RETURNING dan Scan
	query := `INSERT INTO verification_tokens (user_id, token, expired_at) VALUES ($1, $2, $3) RETURNING id, user_id, token, expired_at, created_at`

	err := v.db.QueryRow(ctx, query, verificationToken.UserId, verificationToken.Token, verificationToken.ExpiredAt).
		Scan(&verificationToken.Id, &verificationToken.UserId, &verificationToken.Token, &verificationToken.ExpiredAt, &verificationToken.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("something wrong when create verification token: %w", err)
	}

	return verificationToken, nil
}

func (v *verificationTokenRepository) FindByToken(ctx context.Context, token string) (*domain.VerificationToken, error) {
	var verificationToken domain.VerificationToken
	// Ambil juga kolom created_at
	query := "SELECT id, user_id, token, expired_at, created_at FROM verification_tokens WHERE token = $1"

	err := v.db.QueryRow(ctx, query, token).
		Scan(&verificationToken.Id, &verificationToken.UserId, &verificationToken.Token, &verificationToken.ExpiredAt, &verificationToken.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("verification token not found")
		}
		return nil, fmt.Errorf("something wrong when find verification token by token : %w", err)
	}

	return &verificationToken, nil
}

func (v *verificationTokenRepository) DeleteByUserId(ctx context.Context, userId int64) error {
	_, err := v.db.Exec(ctx, "DELETE FROM verification_tokens WHERE user_id = $1", userId)
	if err != nil {
		return fmt.Errorf("something wrong when delete verification token by user id: %w", err)
	}
	return nil
}
