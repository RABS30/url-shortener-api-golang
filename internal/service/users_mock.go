package service

import (
	"context"
	"shorter-url/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, email, password string) (*domain.User, error) {
	args := m.Called(ctx, email, password)

	var user *domain.User

	if args.Get(0) != nil {
		user = args.Get(0).(*domain.User)
	}
	return user, args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)

	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ChangePassword(ctx context.Context, userId int64, oldPassword, newPassword string) error {
	args := m.Called(ctx, userId, oldPassword, newPassword)
	return args.Error(0)
}

func (m *MockAuthService) ResetPassword(ctx context.Context, newPassword, resetToken string) error {
	args := m.Called(ctx, newPassword, resetToken)
	return args.Error(0)
}
func (m *MockAuthService) LoginWithGoogle(ctx context.Context, info *domain.GoogleUserInfo) (string, error) {
	args := m.Called(ctx, info)

	return args.String(0), args.Error(1)
}
