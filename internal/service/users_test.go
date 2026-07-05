package service

import (
	"context"
	"errors"
	"shorter-url/internal/domain"
	"shorter-url/internal/helper"
	"shorter-url/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func Test_Service_Register_Pass(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	mockHasher := new(helper.MockPasswordHasher)
	secretKey := []byte("super-secret-key")
	svc := NewUserService(mockRepo, secretKey, mockHasher)

	mockRepo.On("FindByEmail", mock.Anything, "new@mail.com").Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(&domain.User{
		Id:    1,
		Email: "new@mail.com",
	}, nil)
	mockHasher.On("Hash", mock.Anything, "password123").Return("hashed_password", nil)

	result, err := svc.Register(context.Background(), "new@mail.com", "password123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.Id)
	mockRepo.AssertExpectations(t)
}

func Test_Service_Register_EmailAlreadyExist(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	mockHasher := new(helper.MockPasswordHasher)
	svc := NewUserService(mockRepo, []byte("secret"), mockHasher)

	mockRepo.On("FindByEmail", mock.Anything, "exist@mail.com").Return(&domain.User{Id: 1, Email: "exist@mail.com"}, nil)

	result, err := svc.Register(context.Background(), "exist@mail.com", "password123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "email already registered")

	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func Test_Service_Login_Pass(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	mockHasher := new(helper.MockPasswordHasher)
	secretKey := []byte("secret-token-key")
	svc := NewUserService(mockRepo, secretKey, mockHasher)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("mypass123"), bcrypt.DefaultCost)

	mockRepo.On("FindByEmail", mock.Anything, "user@mail.com").Return(&domain.User{
		Id:           99,
		Email:        "user@mail.com",
		PasswordHash: string(hashedPassword),
	}, nil)

	mockHasher.On("Compare", mock.Anything, "mypass123", string(hashedPassword)).Return(nil)

	token, err := svc.Login(context.Background(), "user@mail.com", "mypass123")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockRepo.AssertExpectations(t)
}

func Test_Service_Login_WrongPasswordOrEmailNotFound(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	mockHasher := new(helper.MockPasswordHasher)
	svc := NewUserService(mockRepo, []byte("secret"), mockHasher)

	t.Run("Email NotFound", func(t *testing.T) {
		mockRepo.On("FindByEmail", mock.Anything, "notfound@mail.com").Return(nil, nil).Once()

		token, err := svc.Login(context.Background(), "notfound@mail.com", "any-password")
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorContains(t, err, "invalid email or password")
	})

	t.Run("Wrong Password", func(t *testing.T) {
		mockRepo.On("FindByEmail", mock.Anything, "user@mail.com").Return(&domain.User{
			Id:           1,
			Email:        "user@mail.com",
			PasswordHash: "invalid-hash-bukan-bcrypt",
		}, nil).Once()

		mockHasher.On("Compare", mock.Anything, "wrong-password", "invalid-hash-bukan-bcrypt").Return(errors.New("crypto/bcrypt: hashedPassword is not the hash of the given password")).Once()

		token2, err2 := svc.Login(context.Background(), "user@mail.com", "wrong-password")
		assert.Error(t, err2)
		assert.Empty(t, token2)
		assert.ErrorContains(t, err2, "invalid email or password")
	})
}
