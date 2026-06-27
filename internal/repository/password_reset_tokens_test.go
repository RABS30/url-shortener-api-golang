package repository

import (
	"context"
	"fmt"
	"shorter-url/internal/domain"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

func Test_Create_PasswordResetTokens_Pass(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)

	defer mockPool.Close()
	defer func() {
		err := mockPool.ExpectationsWereMet()
		assert.NoError(t, err)
	}()

	ctx := context.Background()
	expiredAt := time.Now()

	inputData := &domain.PasswordResetTokens{
		UserId:    1,
		Token:     "test-token",
		ExpiredAt: expiredAt,
	}

	expectedData := &domain.PasswordResetTokens{
		Id:        10,
		UserId:    inputData.UserId,
		Token:     inputData.Token,
		ExpiredAt: inputData.ExpiredAt,
	}

	queryPattern := `^INSERT INTO password_reset_tokens \(user_id, token, expired_at\) VALUES \(\$1, \$2, \$3\) RETURNING id, user_id, token, expired_at, created_at$`

	mockRow := pgxmock.NewRows([]string{"id", "user_id", "token", "expired_at", "created_at"}).AddRow(
		expectedData.Id,
		expectedData.UserId,
		expectedData.Token,
		expectedData.ExpiredAt,
		expectedData.CreatedAt,
	)

	mockPool.ExpectQuery(queryPattern).WithArgs(inputData.UserId, inputData.Token, inputData.ExpiredAt).WillReturnRows(mockRow)

	repo := NewPasswordResetTokensRepository(mockPool)
	result, err := repo.Create(ctx, inputData)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, result)
}

func Test_Create_PasswordResetTokens_Fail(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)

	defer mockPool.Close()
	defer func() {
		err := mockPool.ExpectationsWereMet()
		assert.NoError(t, err)
	}()

	ctx := context.Background()
	dateExample := time.Now()

	inputData := &domain.PasswordResetTokens{
		UserId:    1,
		Token:     "test-token",
		ExpiredAt: dateExample,
	}

	queryPattern := `^INSERT INTO password_reset_tokens \(user_id, token, expired_at\) VALUES \(\$1, \$2, \$3\) RETURNING id, user_id, token, expired_at, created_at$`

	mockPool.ExpectQuery(queryPattern).WithArgs(inputData.UserId, inputData.Token, inputData.ExpiredAt).WillReturnError(fmt.Errorf("driver: bad connection"))

	repo := NewPasswordResetTokensRepository(mockPool)
	result, err := repo.Create(ctx, inputData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "driver: bad connection")
}

func Test_Delete_PasswordResetTokens_Pass(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)

	defer mockPool.Close()
	defer func() {
		err := mockPool.ExpectationsWereMet()
		assert.NoError(t, err)
	}()

	ctx := context.Background()

	id := int64(1)

	query := `^DELETE FROM password_reset_tokens WHERE id = \$1$`

	mockPool.ExpectExec(query).WithArgs(id).WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := NewPasswordResetTokensRepository(mockPool)
	err = repo.Delete(ctx, id)

	assert.NoError(t, err)
}

func Test_Delete_PasswordResetTokens_Fail(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()
	defer func() {
		err := mockPool.ExpectationsWereMet()
		assert.NoError(t, err)
	}()

	ctx := context.Background()
	id := int64(99)

	query := `^DELETE FROM password_reset_tokens WHERE id = \$1$`

	mockPool.ExpectExec(query).WithArgs(id).WillReturnResult(pgxmock.NewResult("DELETE", 0))

	repo := NewPasswordResetTokensRepository(mockPool)
	err = repo.Delete(ctx, id)

	assert.Error(t, err)
	assert.ErrorContains(t, err, fmt.Sprintf("there is no data deleted, password reset token with ID %d not found", id))
}

func Test_FindByToken_PasswordResetTokens_Pass(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)

	defer mockPool.Close()
	defer func() {
		err := mockPool.ExpectationsWereMet()
		assert.NoError(t, err)
	}()

	ctx := context.Background()

	token := "test-token"
	expiredAt := time.Now()

	expectedData := &domain.PasswordResetTokens{
		Id:        1,
		UserId:    2,
		Token:     token,
		ExpiredAt: expiredAt,
		CreatedAt: time.Now(),
	}

	query := `^SELECT id, user_id, token, expired_at, created_at FROM password_reset_tokens WHERE token = \$1$`

	mockRow := pgxmock.NewRows([]string{"id", "user_id", "token", "expired_at", "created_at"}).AddRow(
		expectedData.Id,
		expectedData.UserId,
		expectedData.Token,
		expectedData.ExpiredAt,
		expectedData.CreatedAt,
	)

	mockPool.ExpectQuery(query).WithArgs(token).WillReturnRows(mockRow)

	repo := NewPasswordResetTokensRepository(mockPool)
	result, err := repo.FindByToken(ctx, token)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, result)
}

func Test_FindByToken_PasswordResetTokens_Fail(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)

	defer mockPool.Close()
	defer func() {
		err := mockPool.ExpectationsWereMet()
		assert.NoError(t, err)
	}()

	ctx := context.Background()
	token := "test-token"

	query := `^SELECT id, user_id, token, expired_at, created_at FROM password_reset_tokens WHERE token = \$1$`

	mockPool.ExpectQuery(query).WithArgs(token).WillReturnError(pgx.ErrNoRows)

	repo := NewPasswordResetTokensRepository(mockPool)
	result, err := repo.FindByToken(ctx, token)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "token not found")
}

func Test_FindByToken_PasswordResetTokens_Error(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)

	defer mockPool.Close()
	defer func() {
		err := mockPool.ExpectationsWereMet()
		assert.NoError(t, err)
	}()

	ctx := context.Background()
	token := "test-token"

	query := `^SELECT id, user_id, token, expired_at, created_at FROM password_reset_tokens WHERE token = \$1$`
	mockPool.ExpectQuery(query).WithArgs(token).WillReturnError(fmt.Errorf("db error"))

	repo := NewPasswordResetTokensRepository(mockPool)
	result, err := repo.FindByToken(ctx, token)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "something wrong when find password reset token by token")
}
