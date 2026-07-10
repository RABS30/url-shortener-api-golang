package repository

import (
	"context"
	"shorter-url/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	args := m.Called(ctx, user)
	var res *domain.User
	if args.Get(0) != nil {
		res = args.Get(0).(*domain.User)
	}
	return res, args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	var user *domain.User
	if args.Get(0) != nil {
		user = args.Get(0).(*domain.User)
	}
	return user, args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	args := m.Called(ctx, user)
	var res *domain.User
	if args.Get(0) != nil {
		res = args.Get(0).(*domain.User)
	}
	return res, args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id int64, hashedPassword string) error {
	args := m.Called(ctx, id, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateVerified(ctx context.Context, id int64, verify bool) error {
	args := m.Called(ctx, id, verify)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) FindById(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	var user *domain.User
	if args.Get(0) != nil {
		user = args.Get(0).(*domain.User)
	}
	return user, args.Error(1)
}

func (m *MockUserRepository) Upsert(ctx context.Context, user *domain.User) (*domain.User, error) {
	args := m.Called(ctx, user)
	var res *domain.User
	if args.Get(0) != nil {
		res = args.Get(0).(*domain.User)
	}
	return res, args.Error(1)
}
