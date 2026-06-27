package helper

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) Hash(ctx context.Context, p string) (string, error) {
	return m.Called(ctx, p).String(0), m.Called(ctx, p).Error(1)
}

func (m *MockPasswordHasher) Compare(ctx context.Context, password string, hashedPassword string) error {
	args := m.Called(ctx, password, hashedPassword)
	return args.Error(0)
}
