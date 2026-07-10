package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"
	"shorter-url/internal/domain"
	"time"
)

type shortUrlsService struct {
	repo domain.ShortUrlsRepository
}

func NewShortUrlService(repo domain.ShortUrlsRepository) domain.ShortUrlsService {
	return &shortUrlsService{
		repo: repo,
	}
}

func (s *shortUrlsService) CreateShortUrl(ctx context.Context, userId int64, originalUrl string, expiredAt time.Time) (*domain.ShortUrl, error) {
	_, err := url.ParseRequestURI(originalUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid URL format: %w", err)
	}

	shortCode, err := generateShortCode(8)
	if err != nil {
		return nil, fmt.Errorf("failed generate short code: %w", err)
	}

	for range 3 {
		inputData := &domain.ShortUrl{
			UserId:      userId,
			ShortCode:   shortCode,
			OriginalUrl: originalUrl,
			ExpiredAt:   expiredAt,
		}

		var result *domain.ShortUrl
		result, err = s.repo.Create(ctx, inputData)
		if err == nil {
			return result, nil
		}

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("failed to create short URL after multiple attempts: %w", err)
}

func (s *shortUrlsService) DeleteShortUrl(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete short url with ID %d: %w", id, err)
	}

	return nil
}

func (s *shortUrlsService) GetShortUrlById(ctx context.Context, id int64) (*domain.ShortUrl, error) {
	result, err := s.repo.FindById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get short url with ID %d: %w", id, err)
	}

	return result, nil
}

func (s *shortUrlsService) GetShortUrlsByUserId(ctx context.Context, userId int64) ([]domain.ShortUrl, error) {
	rows, err := s.repo.FindByUserId(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get short urls by user ID %d: %w", userId, err)
	}

	return rows, nil
}

func (s *shortUrlsService) GetShortUrlByShortCode(ctx context.Context, shortCode string) (*domain.ShortUrl, error) {
	result, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get short url by short code %s: %w", shortCode, err)
	}

	return result, nil
}

func generateShortCode(size int) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, size)

	maxLimit := big.NewInt(int64(len(chars)))

	for i := range result {
		num, err := rand.Int(rand.Reader, maxLimit)
		if err != nil {
			return "", err
		}
		result[i] = chars[num.Int64()]
	}
	return string(result), nil
}
