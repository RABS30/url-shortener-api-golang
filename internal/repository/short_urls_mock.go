package repository

import (
	"context"
	"shorter-url/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockShortUrlsRepository struct {
	mock.Mock
}

func (m *MockShortUrlsRepository) Create(ctx context.Context, shortUrl *domain.ShortUrl) (*domain.ShortUrl, error) {
	// Menggunakan matcher dinamis karena generateShortCode menghasilkan string acak tiap dipanggil
	args := m.Called(ctx, mock.AnythingOfType("*domain.ShortUrl"))
	var res *domain.ShortUrl
	if args.Get(0) != nil {
		res = args.Get(0).(*domain.ShortUrl)
	}
	return res, args.Error(1)
}

func (m *MockShortUrlsRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockShortUrlsRepository) FindById(ctx context.Context, id int64) (*domain.ShortUrl, error) {
	args := m.Called(ctx, id)
	var res *domain.ShortUrl
	if args.Get(0) != nil {
		res = args.Get(0).(*domain.ShortUrl)
	}
	return res, args.Error(1)
}

func (m *MockShortUrlsRepository) FindByUserId(ctx context.Context, userId int64) ([]domain.ShortUrl, error) {
	args := m.Called(ctx, userId)
	var res []domain.ShortUrl
	if args.Get(0) != nil {
		res = args.Get(0).([]domain.ShortUrl)
	}
	return res, args.Error(1)
}

func (m *MockShortUrlsRepository) FindByShortCode(ctx context.Context, shortCode string) (*domain.ShortUrl, error) {
	args := m.Called(ctx, shortCode)
	var res *domain.ShortUrl
	if args.Get(0) != nil {
		res = args.Get(0).(*domain.ShortUrl)
	}
	return res, args.Error(1)
}
