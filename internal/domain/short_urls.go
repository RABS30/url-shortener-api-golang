package domain

import (
	"context"
	"time"
)

type ShortUrl struct {
	Id          int64
	UserId      int64
	ShortCode   string
	OriginalUrl string
	ExpiredAt   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ShortUrlsRepository interface {
	Create(ctx context.Context, shortUrl *ShortUrl) (*ShortUrl, error)
	Delete(ctx context.Context, id int64) error
	FindById(ctx context.Context, id int64) (*ShortUrl, error)
	FindByUserId(ctx context.Context, userId int64) ([]ShortUrl, error)
	FindByShortCode(ctx context.Context, shortCode string) (*ShortUrl, error)
}

type ShortUrlsService interface {
	CreateShortUrl(ctx context.Context, userId int64, originalUrl string, expiredAt time.Time) (*ShortUrl, error)
	DeleteShortUrl(ctx context.Context, id int64) error
	GetShortUrlById(ctx context.Context, id int64) (*ShortUrl, error)
	GetShortUrlsByUserId(ctx context.Context, id int64) ([]ShortUrl, error)
	GetShortUrlByShortCode(ctx context.Context, shortCode string) (*ShortUrl, error)
}
