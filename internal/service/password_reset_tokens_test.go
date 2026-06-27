package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"shorter-url/internal/domain"
	"shorter-url/internal/helper"
	"shorter-url/internal/repository"
	"shorter-url/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_RequestResetPassword_Pass(t *testing.T) {
	mockRepo := new(repository.MockPasswordResetTokensRepository)
	mockUsers := new(repository.MockUserRepository)
	mockEmail := new(helper.MockEmailSender)
	mockHasher := new(helper.MockPasswordHasher)

	ctx := context.Background()
	targetEmail := "rabs@example.com"
	existingUser := &domain.User{Id: 1, Email: targetEmail}

	mockUsers.On("FindByEmail", ctx, targetEmail).Return(existingUser, nil)
	mockRepo.On("Create", ctx, mock.Anything).Return(&domain.PasswordResetTokens{}, nil)
	mockEmail.On("SendEmail", ctx, targetEmail, mock.Anything, mock.Anything).Return(nil)

	s := service.NewPasswordResetTokensService(mockRepo, mockUsers, mockEmail, mockHasher, "http://localhost:8080")
	err := s.RequestResetPassword(ctx, targetEmail)

	assert.NoError(t, err)
	mockUsers.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockEmail.AssertExpectations(t)
}

func Test_RequestResetPassword_Fail_UserNotFound(t *testing.T) {
	mockRepo := new(repository.MockPasswordResetTokensRepository)
	mockUsers := new(repository.MockUserRepository)
	mockEmail := new(helper.MockEmailSender)
	mockHasher := new(helper.MockPasswordHasher)

	ctx := context.Background()
	targetEmail := "unknown@example.com"

	mockUsers.On("FindByEmail", ctx, targetEmail).Return(nil, nil)

	s := service.NewPasswordResetTokensService(mockRepo, mockUsers, mockEmail, mockHasher, "http://localhost:8080")
	err := s.RequestResetPassword(ctx, targetEmail)

	assert.NoError(t, err)
	mockUsers.AssertExpectations(t)
}

func Test_RequestResetPassword_Fail_EmailSenderError(t *testing.T) {
	mockRepo := new(repository.MockPasswordResetTokensRepository)
	mockUsers := new(repository.MockUserRepository)
	mockEmail := new(helper.MockEmailSender)
	mockHasher := new(helper.MockPasswordHasher)

	ctx := context.Background()
	targetEmail := "rabs@example.com"
	existingUser := &domain.User{Id: 1, Email: targetEmail}

	mockUsers.On("FindByEmail", ctx, targetEmail).Return(existingUser, nil)
	mockRepo.On("Create", ctx, mock.Anything).Return(&domain.PasswordResetTokens{}, nil)
	mockEmail.On("SendEmail", ctx, targetEmail, mock.Anything, mock.Anything).Return(errors.New("smtp gateway timeout"))

	s := service.NewPasswordResetTokensService(mockRepo, mockUsers, mockEmail, mockHasher, "http://localhost:8080")
	err := s.RequestResetPassword(ctx, targetEmail)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send email")
	mockEmail.AssertExpectations(t)
}

func Test_ExecuteResetPassword_Pass(t *testing.T) {
	mockRepo := new(repository.MockPasswordResetTokensRepository)
	mockUsers := new(repository.MockUserRepository)
	mockEmail := new(helper.MockEmailSender)
	mockHasher := new(helper.MockPasswordHasher)

	ctx := context.Background()
	token := "valid-token-rabs"
	dbToken := &domain.PasswordResetTokens{UserId: 1, Token: token, ExpiredAt: time.Now().Add(time.Minute * 10)}

	mockRepo.On("FindByToken", ctx, token).Return(dbToken, nil)
	mockHasher.On("Hash", ctx, "Secret123").Return("hashed_secret", nil)
	mockUsers.On("UpdatePassword", ctx, int64(1), "hashed_secret").Return(nil)
	mockRepo.On("DeleteByUserId", ctx, int64(1)).Return(nil)

	s := service.NewPasswordResetTokensService(mockRepo, mockUsers, mockEmail, mockHasher, "http://localhost:8080")
	err := s.ExecuteResetPassword(ctx, token, "Secret123", "Secret123")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockHasher.AssertExpectations(t)
	mockUsers.AssertExpectations(t)
}

func Test_ExecuteResetPassword_Fail_PasswordNotEqual(t *testing.T) {
	mockRepo := new(repository.MockPasswordResetTokensRepository)
	mockUsers := new(repository.MockUserRepository)
	mockEmail := new(helper.MockEmailSender)
	mockHasher := new(helper.MockPasswordHasher)

	ctx := context.Background()
	token := "valid-token-rabs"
	dbToken := &domain.PasswordResetTokens{UserId: 1, Token: token, ExpiredAt: time.Now().Add(time.Minute * 10)}

	mockRepo.On("FindByToken", ctx, token).Return(dbToken, nil)

	s := service.NewPasswordResetTokensService(mockRepo, mockUsers, mockEmail, mockHasher, "http://localhost:8080")
	err := s.ExecuteResetPassword(ctx, token, "Secret123", "PasswordBeda")

	assert.Error(t, err)
	assert.Equal(t, "password not equal", err.Error())
}

func Test_ExecuteResetPassword_Fail_TokenExpired(t *testing.T) {
	mockRepo := new(repository.MockPasswordResetTokensRepository)
	mockUsers := new(repository.MockUserRepository)
	mockEmail := new(helper.MockEmailSender)
	mockHasher := new(helper.MockPasswordHasher)

	ctx := context.Background()
	token := "expired-token-rabs"
	dbToken := &domain.PasswordResetTokens{UserId: 1, Token: token, ExpiredAt: time.Now().Add(-time.Minute * 5)}

	mockRepo.On("FindByToken", ctx, token).Return(dbToken, nil)

	s := service.NewPasswordResetTokensService(mockRepo, mockUsers, mockEmail, mockHasher, "http://localhost:8080")
	err := s.ExecuteResetPassword(ctx, token, "Secret123", "Secret123")

	assert.Error(t, err)
	assert.Equal(t, "token expired", err.Error())
}
