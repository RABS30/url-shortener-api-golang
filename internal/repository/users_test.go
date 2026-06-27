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

func Test_Create_Users_Pass(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()
	defer func() {
		assert.NoError(t, mockPool.ExpectationsWereMet())
	}()

	ctx := context.Background()
	inputData := &domain.User{
		Email:        "contoh@gmail.com",
		PasswordHash: "iniadalahhashrandom",
	}

	id := int64(9)
	isVerified := false
	status := "active"
	fixedTime := time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC)

	query := `^INSERT INTO users\(email, password_hash\)VALUES\(\$1, \$2\) RETURNING id, email, password_hash, is_verified, status, created_at$`

	mockRow := pgxmock.NewRows([]string{"id", "email", "password_hash", "is_verified", "status", "created_at"}).
		AddRow(id, inputData.Email, inputData.PasswordHash, isVerified, status, fixedTime)

	mockPool.ExpectQuery(query).WithArgs(inputData.Email, inputData.PasswordHash).WillReturnRows(mockRow)

	repo := NewUserRepository(mockPool)
	result, err := repo.Create(ctx, inputData)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, id, result.Id)
	assert.Equal(t, fixedTime, result.CreatedAt)
}

func Test_Create_Duplicate_Users(t *testing.T) {
	mockpool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockpool.Close()
	defer func() {
		assert.NoError(t, mockpool.ExpectationsWereMet())
	}()

	ctx := context.Background()
	inputData := &domain.User{
		Email:        "contoh@gmail.com",
		PasswordHash: "iniadalahpasswordhash",
	}

	query := `^INSERT INTO users\(email, password_hash\)VALUES\(\$1, \$2\) RETURNING id, email, password_hash, is_verified, status, created_at$`

	mockpool.ExpectQuery(query).WithArgs(inputData.Email, inputData.PasswordHash).
		WillReturnError(fmt.Errorf("ERROR: Duplicate key value violates unique constraint (SQLSTATE 23505)"))

	repo := NewUserRepository(mockpool)
	result, err := repo.Create(ctx, inputData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "something wrong when create new data")
	assert.ErrorContains(t, err, "23505")
}

func Test_Update_Users_Pass(t *testing.T) {
	mockpool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockpool.Close()
	defer func() {
		assert.NoError(t, mockpool.ExpectationsWereMet())
	}()

	ctx := context.Background()
	inputData := &domain.User{
		Id:           9,
		PasswordHash: "iniadalahpasswordhash",
		IsVerified:   true,
		Status:       "active",
	}
	fixedTime := time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC)
	query := `^UPDATE users SET password_hash = \$1, is_verified = \$2, status = \$3 WHERE id = \$4 RETURNING id, email, password_hash, is_verified, status, created_at$`

	mockRow := pgxmock.NewRows([]string{"id", "email", "password_hash", "is_verified", "status", "created_at"}).
		AddRow(inputData.Id, "tester@gmail.com", inputData.PasswordHash, inputData.IsVerified, inputData.Status, fixedTime)

	mockpool.ExpectQuery(query).WithArgs(inputData.PasswordHash, inputData.IsVerified, inputData.Status, inputData.Id).WillReturnRows(mockRow)

	repo := NewUserRepository(mockpool)
	result, err := repo.Update(ctx, inputData)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, fixedTime, result.CreatedAt)
}

func Test_Update_Users_Fail(t *testing.T) {
	mockpool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockpool.Close()
	defer func() {
		assert.NoError(t, mockpool.ExpectationsWereMet())
	}()

	ctx := context.Background()
	inputData := &domain.User{
		Id:           1,
		PasswordHash: "passwordhashhash",
		IsVerified:   false,
		Status:       "inactive",
	}

	query := `^UPDATE users SET password_hash = \$1, is_verified = \$2, status = \$3 WHERE id = \$4 RETURNING id, email, password_hash, is_verified, status, created_at$`

	mockpool.ExpectQuery(query).WithArgs(inputData.PasswordHash, inputData.IsVerified, inputData.Status, inputData.Id).WillReturnError(pgx.ErrNoRows)

	repo := NewUserRepository(mockpool)
	result, err := repo.Update(ctx, inputData)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.ErrorContains(t, err, fmt.Sprintf("user dengan ID %d tidak ditemukan", inputData.Id))
}

func Test_Delete_Users_Pass(t *testing.T) {
	mockpool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockpool.Close()
	defer func() {
		assert.NoError(t, mockpool.ExpectationsWereMet())
	}()

	ctx := context.Background()
	id := int64(1)
	query := `^DELETE FROM users WHERE id = \$1$`

	mockpool.ExpectExec(query).WithArgs(id).WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := NewUserRepository(mockpool)
	err = repo.Delete(ctx, id)

	assert.NoError(t, err)
}

func Test_Delete_Users_Fail(t *testing.T) {
	mockpool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockpool.Close()
	defer func() {
		assert.NoError(t, mockpool.ExpectationsWereMet())
	}()

	ctx := context.Background()
	id := int64(99)
	query := `^DELETE FROM users WHERE id = \$1$`

	mockpool.ExpectExec(query).WithArgs(id).WillReturnResult(pgxmock.NewResult("DELETE", 0))

	repo := NewUserRepository(mockpool)
	err = repo.Delete(ctx, id)

	assert.ErrorContains(t, err, "there is no data deleted")
}

func Test_FindById_Users_Pass(t *testing.T) {
	mockpool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockpool.Close()
	defer func() {
		assert.NoError(t, mockpool.ExpectationsWereMet())
	}()

	ctx := context.Background()
	id := int64(1)
	fixedTime := time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC)

	expectedData := &domain.User{
		Id:           id,
		Email:        "rokubi27@gmail.com",
		PasswordHash: "this_password_is_hashing",
		IsVerified:   true,
		Status:       "active",
		CreatedAt:    fixedTime,
	}

	query := `^SELECT id, email, password_hash, is_verified, status, created_at FROM users WHERE id = \$1$`

	mockRow := pgxmock.NewRows([]string{"id", "email", "password_hash", "is_verified", "status", "created_at"}).
		AddRow(expectedData.Id, expectedData.Email, expectedData.PasswordHash, expectedData.IsVerified, expectedData.Status, expectedData.CreatedAt)

	mockpool.ExpectQuery(query).WithArgs(id).WillReturnRows(mockRow)

	repo := NewUserRepository(mockpool)
	result, err := repo.FindById(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, result)
}

func Test_FindById_Users_Fail(t *testing.T) {
	mockpool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockpool.Close()
	defer func() {
		assert.NoError(t, mockpool.ExpectationsWereMet())
	}()

	ctx := context.Background()
	id := int64(99)
	query := `^SELECT id, email, password_hash, is_verified, status, created_at FROM users WHERE id = \$1$`

	mockpool.ExpectQuery(query).WithArgs(id).WillReturnError(pgx.ErrNoRows)

	repo := NewUserRepository(mockpool)
	result, err := repo.FindById(ctx, id)

	assert.Nil(t, result)
	assert.ErrorContains(t, err, "user dengan ID 99 tidak ditemukan")
}

func Test_FindByEmail_Users_Pass(t *testing.T) {
	mockpool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockpool.Close()
	defer func() {
		assert.NoError(t, mockpool.ExpectationsWereMet())
	}()

	ctx := context.Background()
	email := "rokubi27@gmail.com"
	fixedTime := time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC)

	expectedData := &domain.User{
		Id:           1,
		Email:        email,
		PasswordHash: "this_password_is_hashing",
		IsVerified:   true,
		Status:       "active",
		CreatedAt:    fixedTime,
	}

	query := `^SELECT id, email, password_hash, is_verified, status, created_at FROM users WHERE email = \$1$`

	mockRow := pgxmock.NewRows([]string{"id", "email", "password_hash", "is_verified", "status", "created_at"}).
		AddRow(expectedData.Id, expectedData.Email, expectedData.PasswordHash, expectedData.IsVerified, expectedData.Status, expectedData.CreatedAt)

	mockpool.ExpectQuery(query).WithArgs(email).WillReturnRows(mockRow)

	repo := NewUserRepository(mockpool)
	result, err := repo.FindByEmail(ctx, email)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, result)
}

func Test_FindByEmail_Users_Fail(t *testing.T) {
	mockpool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockpool.Close()
	defer func() {
		assert.NoError(t, mockpool.ExpectationsWereMet())
	}()

	ctx := context.Background()
	email := "rokubi27@gmail.com"
	query := `^SELECT id, email, password_hash, is_verified, status, created_at FROM users WHERE email = \$1$`

	mockpool.ExpectQuery(query).WithArgs(email).WillReturnError(pgx.ErrNoRows)

	repo := NewUserRepository(mockpool)
	result, err := repo.FindByEmail(ctx, email)

	assert.Nil(t, result)
	assert.ErrorContains(t, err, "user not found")
}
