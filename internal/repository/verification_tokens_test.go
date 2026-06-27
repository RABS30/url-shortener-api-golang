package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"shorter-url/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

func Test_Create_VerificationTokens_Pass(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()
	defer func() {
		assert.NoError(t, mockPool.ExpectationsWereMet())
	}()

	ctx := context.Background()

	fixedTime := time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC)

	inputData := &domain.VerificationToken{
		UserId:    1,
		Token:     "test-token",
		ExpiredAt: fixedTime,
	}

	expectedData := &domain.VerificationToken{
		Id:        10,
		UserId:    inputData.UserId,
		Token:     inputData.Token,
		ExpiredAt: inputData.ExpiredAt,
		CreatedAt: time.Now(),
	}

	queryPattern := `^INSERT INTO verification_tokens \(user_id, token, expired_at\) VALUES \(\$1, \$2, \$3\) RETURNING id, user_id, token, expired_at, created_at$`

	mockRow := pgxmock.NewRows([]string{"id", "user_id", "token", "expired_at", "created_at"}).AddRow(
		expectedData.Id,
		expectedData.UserId,
		expectedData.Token,
		expectedData.ExpiredAt,
		expectedData.CreatedAt,
	)

	mockPool.ExpectQuery(queryPattern).
		WithArgs(inputData.UserId, inputData.Token, inputData.ExpiredAt).
		WillReturnRows(mockRow)

	repo := NewVerificationTokenRepository(mockPool)
	result, err := repo.Create(ctx, inputData)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, result)
}

func Test_Create_VerificationTokens_Fail(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()
	defer func() {
		assert.NoError(t, mockPool.ExpectationsWereMet())
	}()

	ctx := context.Background()
	fixedTime := time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC)

	inputData := &domain.VerificationToken{
		UserId:    1,
		Token:     "test-token",
		ExpiredAt: fixedTime,
	}

	queryPattern := `^INSERT INTO verification_tokens \(user_id, token, expired_at\) VALUES \(\$1, \$2, \$3\) RETURNING id, user_id, token, expired_at, created_at$`

	mockPool.ExpectQuery(queryPattern).
		WithArgs(inputData.UserId, inputData.Token, inputData.ExpiredAt).
		WillReturnError(fmt.Errorf("db error"))

	repo := NewVerificationTokenRepository(mockPool)
	result, err := repo.Create(ctx, inputData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "something wrong when create verification token")
}

func Test_Delete_VerificationTokens_Pass(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()
	defer func() {
		assert.NoError(t, mockPool.ExpectationsWereMet())
	}()

	ctx := context.Background()
	id := int64(1)

	query := `^DELETE FROM verification_tokens WHERE id = \$1$`

	mockPool.ExpectExec(query).WithArgs(id).WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := NewVerificationTokenRepository(mockPool)
	err = repo.Delete(ctx, id)

	assert.NoError(t, err)
}

func Test_Delete_VerificationTokens_Fail(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()
	defer func() {
		assert.NoError(t, mockPool.ExpectationsWereMet())
	}()

	ctx := context.Background()
	id := int64(99)

	query := `^DELETE FROM verification_tokens WHERE id = \$1$`

	// Menghasilkan RowsAffected = 0 untuk simulasi data tidak ketemu saat DELETE
	mockPool.ExpectExec(query).WithArgs(id).WillReturnResult(pgxmock.NewResult("DELETE", 0))

	repo := NewVerificationTokenRepository(mockPool)
	err = repo.Delete(ctx, id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification token with id 99 not found")
}

func Test_FindByToken_VerificationTokens_Pass(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()
	defer func() {
		assert.NoError(t, mockPool.ExpectationsWereMet())
	}()

	ctx := context.Background()
	token := "test-token"
	fixedTime := time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC)

	expectedData := &domain.VerificationToken{
		Id:        1,
		UserId:    2,
		Token:     token,
		ExpiredAt: fixedTime,
		CreatedAt: time.Now(),
	}

	query := `^SELECT id, user_id, token, expired_at, created_at FROM verification_tokens WHERE token = \$1$`

	mockRow := pgxmock.NewRows([]string{"id", "user_id", "token", "expired_at", "created_at"}).AddRow(
		expectedData.Id,
		expectedData.UserId,
		expectedData.Token,
		expectedData.ExpiredAt,
		expectedData.CreatedAt,
	)

	mockPool.ExpectQuery(query).WithArgs(token).WillReturnRows(mockRow)

	repo := NewVerificationTokenRepository(mockPool)
	result, err := repo.FindByToken(ctx, token)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, result)
}

func Test_FindByToken_VerificationTokens_Fail(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()
	defer func() {
		assert.NoError(t, mockPool.ExpectationsWereMet())
	}()

	ctx := context.Background()
	token := "test-token"

	query := `^SELECT id, user_id, token, expired_at, created_at FROM verification_tokens WHERE token = \$1$`

	// QueryRow().Scan() data kosong mengembalikan pgx.ErrNoRows asli database
	mockPool.ExpectQuery(query).WithArgs(token).WillReturnError(pgx.ErrNoRows)

	repo := NewVerificationTokenRepository(mockPool)
	result, err := repo.FindByToken(ctx, token)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "verification token not found")
}

func Test_FindByToken_VerificationTokens_Error(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()
	defer func() {
		assert.NoError(t, mockPool.ExpectationsWereMet())
	}()

	ctx := context.Background()
	token := "test-token"

	query := `^SELECT id, user_id, token, expired_at, created_at FROM verification_tokens WHERE token = \$1$`

	mockPool.ExpectQuery(query).WithArgs(token).WillReturnError(fmt.Errorf("db error"))

	repo := NewVerificationTokenRepository(mockPool)
	result, err := repo.FindByToken(ctx, token)

	assert.Error(t, err)
	assert.Nil(t, result)

	assert.Contains(t, err.Error(), "something wrong when find verification token by token ")
}
