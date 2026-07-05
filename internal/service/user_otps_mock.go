package service

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockUserOtpsService struct {
	mock.Mock
}

func (m *MockUserOtpsService) SendOTP(ctx context.Context, email string, otpType string) error {
	args := m.Called(ctx, email, otpType)
	return args.Error(0)
}

func (m *MockUserOtpsService) VerifyOTP(ctx context.Context, code string, email string, otpType string) (string, error) {
	args := m.Called(ctx, code, email, otpType)
	return args.String(0), args.Error(1)
}
