package repository

import (
	"context"
	"errors"
	"fmt"
	"shorter-url/internal/database"
	"shorter-url/internal/domain"

	"github.com/jackc/pgx/v5"
)

type userOtpsRepository struct {
	db database.PgxDatabase
}

func NewUserOtpsRepository(db database.PgxDatabase) domain.UserOtpsRepository {
	return &userOtpsRepository{
		db: db,
	}
}

func (r *userOtpsRepository) Create(ctx context.Context, userOtps *domain.UserOtps) (*domain.UserOtps, error) {
	query := `INSERT INTO user_otps(email, otp_code, type, expired_at ) VALUES($1, $2, $3, $4) RETURNING id, email, otp_code, type, is_used, expired_at, created_at`

	err := r.db.QueryRow(ctx, query, userOtps.Email, userOtps.Code, userOtps.Type, userOtps.ExpiredAt).Scan(&userOtps.Id, &userOtps.Email, &userOtps.Code, &userOtps.Type, &userOtps.IsUsed, &userOtps.ExpiredAt, &userOtps.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert otp code: %w", err)
	}

	return userOtps, nil
}

func (r *userOtpsRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM user_otps WHERE id = $1`

	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user otps by id: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("delete user otps by id: %w", domain.ErrNoRows)

	}

	return nil
}
func (r *userOtpsRepository) DeleteByCodeOTP(ctx context.Context, code string) error {
	query := `DELETE FROM user_otps WHERE otp_code = $1`

	commandTag, err := r.db.Exec(ctx, query, code)
	if err != nil {
		return fmt.Errorf("delete user otps by id: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("delete user otps by id: %w", domain.ErrNoRows)

	}

	return nil
}

func (r *userOtpsRepository) UpdateIsUsed(ctx context.Context, id int64, isUsed bool) error {
	query := `UPDATE user_otps SET is_used = $1 WHERE id = $2`

	commandTag, err := r.db.Exec(ctx, query, isUsed, id)
	if err != nil {
		return fmt.Errorf("update is_used user otps: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("update id_used user otps: %w", domain.ErrNotFound)
	}

	return nil
}

func (r *userOtpsRepository) FindByEmailCodeType(ctx context.Context, email string, code string, otpType string) (*domain.UserOtps, error) {
	query := `SELECT id, email, otp_code, type, is_used, expired_at, created_at FROM user_otps WHERE email = $1 AND otp_code = $2 AND type = $3`

	result := &domain.UserOtps{}

	err := r.db.QueryRow(ctx, query, email, code, otpType).Scan(&result.Id, &result.Email, &result.Code, &result.Type, &result.IsUsed, &result.ExpiredAt, &result.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = domain.ErrNotFound
		}
		return nil, fmt.Errorf("query user otps by email, code, type: %w", err)
	}
	return result, nil
}

func (r *userOtpsRepository) FindById(ctx context.Context, id int64) (*domain.UserOtps, error) {
	query := `SELECT id, email, otp_code, type, is_used, expired_at, created_at FROM user_otps WHERE id = $1`

	result := &domain.UserOtps{}

	err := r.db.QueryRow(ctx, query, id).Scan(&result.Id, &result.Email, &result.Code, &result.Type, &result.IsUsed, &result.ExpiredAt, &result.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = domain.ErrNotFound
		}
		return nil, fmt.Errorf("query find by id: %w", err)
	}

	return result, nil
}
