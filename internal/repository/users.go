package repository

import (
	"context"
	"errors"
	"fmt"
	"shorter-url/internal/database"
	"shorter-url/internal/domain"

	"github.com/jackc/pgx/v5"
)

type userRepository struct {
	db database.PgxDatabase
}

func NewUserRepository(db database.PgxDatabase) domain.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (u *userRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `INSERT INTO users(email, password_hash)VALUES($1, $2) RETURNING id, email, password_hash, is_verified, status, created_at`

	err := u.db.QueryRow(ctx, query, user.Email, user.PasswordHash).Scan(&user.Id, &user.Email, &user.PasswordHash, &user.IsVerified, &user.Status, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("something wrong when create new data: %w", err)
	}

	return user, nil
}

func (u *userRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `UPDATE users SET password_hash = $1, is_verified = $2, status = $3 WHERE id = $4 RETURNING id, email, password_hash, is_verified, status, created_at`

	err := u.db.QueryRow(ctx, query, user.PasswordHash, user.IsVerified, user.Status, user.Id).
		Scan(&user.Id, &user.Email, &user.PasswordHash, &user.IsVerified, &user.Status, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user dengan ID %d tidak ditemukan", user.Id)
		}
		return nil, fmt.Errorf("something wrong when update data: %w", err)
	}

	return user, nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, userId int64, hashedPassword string) error {
	query := "UPDATE users SET password_hash = $1 WHERE id = $2"

	result, err := r.db.Exec(ctx, query, hashedPassword, userId)
	if err != nil {
		return fmt.Errorf("something wrong when update user password: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found, no rows updated")
	}

	return nil
}

func (u *userRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	commandTag, err := u.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("something wrong when delete data : %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("there is no data deleted, user with ID %d not found", id)
	}
	return nil
}

func (u *userRepository) FindById(ctx context.Context, id int64) (*domain.User, error) {
	query := `SELECT id, email, password_hash, is_verified, status, created_at FROM users WHERE id = $1`
	user := &domain.User{}

	err := u.db.QueryRow(ctx, query, id).Scan(&user.Id, &user.Email, &user.PasswordHash, &user.IsVerified, &user.Status, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user dengan ID %d tidak ditemukan", id)
		}
		return nil, fmt.Errorf("something error when find user by id : %w", err)
	}
	return user, nil
}

func (u *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, is_verified, status, created_at FROM users WHERE email = $1`
	user := &domain.User{}

	err := u.db.QueryRow(ctx, query, email).Scan(&user.Id, &user.Email, &user.PasswordHash, &user.IsVerified, &user.Status, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found, %w", err)
		}
		return nil, fmt.Errorf("something wrong when find short url by id : %w", err)
	}
	return user, nil
}
