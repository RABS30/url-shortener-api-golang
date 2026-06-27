package repository

import (
	"context"
	"shorter-url/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockPasswordResetTokensRepository adalah implementasi mock penuh dari domain.PasswordResetTokensRepository rabs
type MockPasswordResetTokensRepository struct {
	mock.Mock
}

func NewMockPasswordResetTokensRepository() *MockPasswordResetTokensRepository {
	return &MockPasswordResetTokensRepository{}
}

func (m *MockPasswordResetTokensRepository) Create(ctx context.Context, passwordResetToken *domain.PasswordResetTokens) (*domain.PasswordResetTokens, error) {
	args := m.Called(ctx, passwordResetToken)

	if args.Get(0) != nil {
		return args.Get(0).(*domain.PasswordResetTokens), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockPasswordResetTokensRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPasswordResetTokensRepository) DeleteByUserId(ctx context.Context, userId int64) error {
	args := m.Called(ctx, userId)
	return args.Error(0)
}

func (m *MockPasswordResetTokensRepository) FindByToken(ctx context.Context, token string) (*domain.PasswordResetTokens, error) {
	args := m.Called(ctx, token)

	if args.Get(0) != nil {
		return args.Get(0).(*domain.PasswordResetTokens), args.Error(1)
	}

	return nil, args.Error(1)
}
