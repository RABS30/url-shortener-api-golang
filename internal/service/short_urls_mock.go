package service

import (
	"context"
	"shorter-url/internal/domain"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockShortUrlService struct {
	mock.Mock
}

func (m *MockShortUrlService) CreateShortUrl(ctx context.Context, userId int64, originalUrl string, expiredAt time.Time) (*domain.ShortUrl, error) {
	args := m.Called(ctx, userId, originalUrl, expiredAt)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ShortUrl), args.Error(1)
}

func (m *MockShortUrlService) DeleteShortUrl(ctx context.Context, id int64) error {
	return nil
}
func (m *MockShortUrlService) GetShortUrlById(ctx context.Context, id int64) (*domain.ShortUrl, error) {
	return nil, nil
}
func (m *MockShortUrlService) GetShortUrlsByUserId(ctx context.Context, id int64) ([]domain.ShortUrl, error) {
	return nil, nil
}
func (m *MockShortUrlService) GetShortUrlByShortCode(ctx context.Context, shortCode string) (*domain.ShortUrl, error) {
	args := m.Called(ctx, shortCode)

	var result *domain.ShortUrl
	if args.Get(0) != nil {
		result = args.Get(0).(*domain.ShortUrl)
	}

	return result, args.Error(1)
}
