package service

import (
	"context"
	"fmt"
	"shorter-url/internal/domain"
)

type clickEventService struct {
	repo domain.ClickEventRepository
}

func NewClickEventService(repo domain.ClickEventRepository) domain.ClickEventService {
	return &clickEventService{
		repo: repo,
	}
}

func (s *clickEventService) Create(ctx context.Context, clickEvent *domain.ClickEvent) (*domain.ClickEvent, error) {
	result, err := s.repo.Create(ctx, clickEvent)
	if err != nil {
		return nil, fmt.Errorf("something error when create click event, %w", err)
	}

	return result, nil
}

func (s *clickEventService) FindByShortUrlId(ctx context.Context, shortUrlId int64, userId int64) ([]domain.ClickEvent, error) {
	// Oper userID ke layer repository rabs
	listEvent, err := s.repo.FindByShortUrlId(ctx, shortUrlId, userId)
	if err != nil {
		return nil, fmt.Errorf("something error when get list click event, %w", err)
	}

	return listEvent, nil
}
