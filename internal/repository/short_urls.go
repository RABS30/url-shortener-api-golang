package repository

import (
	"context"
	"fmt"
	"shorter-url/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type shortUrlRepository struct {
	db *pgxpool.Pool
}

func NewShortUrlRepository(db *pgxpool.Pool) domain.ShortUrlsRepository {
	return &shortUrlRepository{
		db: db,
	}
}

func (r *shortUrlRepository) Create(ctx context.Context, shortUrl *domain.ShortUrl) (*domain.ShortUrl, error) {
	query := `INSERT INTO short_urls (user_id, original_url, short_code, expired_at) VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRow(ctx, query, shortUrl.UserId, shortUrl.OriginalUrl, shortUrl.ShortCode, shortUrl.ExpiredAt).Scan(&shortUrl.Id)
	if err != nil {
		return nil, fmt.Errorf("something wrong when create short url : %w", err)
	}
	return shortUrl, nil
}

func (r *shortUrlRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM short_urls WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("something wrong when delete short url : %w", err)
	}
	return nil
}

func (r *shortUrlRepository) FindById(ctx context.Context, id int64) (*domain.ShortUrl, error) {
	query := `SELECT id, user_id, short_code, original_url, expired_at, created_at, updated_at FROM short_urls WHERE id = $1`
	shortUrl := &domain.ShortUrl{}
	err := r.db.QueryRow(ctx, query, id).Scan(&shortUrl.Id, &shortUrl.UserId, &shortUrl.ShortCode, &shortUrl.OriginalUrl, &shortUrl.ExpiredAt, &shortUrl.CreatedAt, &shortUrl.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("something wrong when find short url by id : %w", err)
	}
	return shortUrl, nil
}

func (r *shortUrlRepository) FindByUserId(ctx context.Context, userId int64) ([]domain.ShortUrl, error) {
	query := `SELECT id, user_id, short_code, original_url, expired_at, created_at, updated_at FROM short_urls WHERE user_id = $1`

	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("something wrong when find short url by user id : %w", err)
	}

	defer rows.Close()

	var shortUrls []domain.ShortUrl

	for rows.Next() {
		var shortUrl domain.ShortUrl

		err := rows.Scan(&shortUrl.Id, &shortUrl.UserId, &shortUrl.ShortCode, &shortUrl.OriginalUrl, &shortUrl.ExpiredAt, &shortUrl.CreatedAt, &shortUrl.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("something wrong when scan short url : %w", err)
		}

		shortUrls = append(shortUrls, shortUrl)
	}

	return shortUrls, nil
}

func (r *shortUrlRepository) FindByShortCode(ctx context.Context, shortCode string) (*domain.ShortUrl, error) {
	query := `SELECT id, user_id, short_code, original_url, expired_at, created_at, updated_at FROM short_urls WHERE short_code = $1`
	shortUrl := &domain.ShortUrl{}
	err := r.db.QueryRow(ctx, query, shortCode).Scan(&shortUrl.Id, &shortUrl.UserId, &shortUrl.ShortCode, &shortUrl.OriginalUrl, &shortUrl.ExpiredAt, &shortUrl.CreatedAt, &shortUrl.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("something wrong when find short url by short code : %w", err)
	}
	return shortUrl, nil
}
