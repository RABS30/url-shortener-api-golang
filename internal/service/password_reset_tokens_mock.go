package service

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockPasswordResetTokensService struct {
	mock.Mock
}

func (m *MockPasswordResetTokensService) RequestResetPassword(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockPasswordResetTokensService) ExecuteResetPassword(ctx context.Context, token, pass1, pass2 string) error {
	args := m.Called(ctx, token, pass1, pass2)
	return args.Error(0)
}
