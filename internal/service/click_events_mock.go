package service

import (
	"context"
	"shorter-url/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockClickEventService struct {
	mock.Mock
}

func (m *MockClickEventService) Create(ctx context.Context, clickEvent *domain.ClickEvent) (*domain.ClickEvent, error) {
	args := m.Called(ctx, clickEvent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ClickEvent), args.Error(1)
}

func (m *MockClickEventService) FindByShortUrlId(ctx context.Context, shortUrlId int64, userId int64) ([]domain.ClickEvent, error) {
	args := m.Called(ctx, shortUrlId, userId)

	var clickEvents []domain.ClickEvent

	if args.Get(0) != nil {
		clickEvents = args.Get(0).([]domain.ClickEvent)
	}

	return clickEvents, args.Error(1)
}
