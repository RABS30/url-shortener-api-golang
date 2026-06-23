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
		return nil, fmt.Errorf("URL format not valid: %w", err)
	}

	for range 3 {
		inputData := &domain.ShortUrl{
			UserId:      userId,
			ShortCode:   generateShortCode(),
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

	return nil, fmt.Errorf("Failed to create short URL. Please recreate. %w", err)
}

func (s *shortUrlsService) DeleteShortUrl(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to Delete with ID %d. %w", id, err)
	}

	return nil
}

func (s *shortUrlsService) GetShortUrlById(ctx context.Context, id int64) (*domain.ShortUrl, error) {
	result, err := s.repo.FindById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get short url with ID %d. : %w", id, err)
	}

	return result, nil
}

func (s *shortUrlsService) GetShortUrlsByUserId(ctx context.Context, userId int64) ([]domain.ShortUrl, error) {
	rows, err := s.repo.FindByUserId(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get data by user id %d, %w", userId, err)
	}

	return rows, nil
}

func (s *shortUrlsService) GetShortUrlByShortCode(ctx context.Context, shortCode string) (*domain.ShortUrl, error) {
	result, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get data by short code %s, %w ", shortCode, err)
	}

	return result, nil
}

func generateShortCode() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 6) 
	
	maxLimit := big.NewInt(int64(len(chars)))

	for i := range result {
		num, _ := rand.Int(rand.Reader, maxLimit)
		result[i] = chars[num.Int64()]
	}
	return string(result)
}
