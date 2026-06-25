package repository

import (
	"context"
	"fmt"
	"shorter-url/internal/database"
	"shorter-url/internal/domain"
)

type clickEventsRepository struct {
	db database.PgxDatabase
}

func NewClickEventsRepository(db database.PgxDatabase) domain.ClickEventRepository {
	return &clickEventsRepository{
		db: db,
	}
}

func (r *clickEventsRepository) Create(ctx context.Context, clickEvent *domain.ClickEvent) (*domain.ClickEvent, error) {
	query := `INSERT INTO click_events(short_url_id, ip_address, user_agent, referer)VALUES($1, $2, $3, $4) RETURNING id, ip_address, short_url_id, user_agent, referer, clicked_at`

	err := r.db.QueryRow(ctx, query, clickEvent.ShortUrlId, clickEvent.IpAddress, clickEvent.UserAgent, clickEvent.Referer).Scan(&clickEvent.Id, &clickEvent.IpAddress, &clickEvent.ShortUrlId, &clickEvent.UserAgent, &clickEvent.Referer, &clickEvent.ClickedAt)
	if err != nil {
		return nil, fmt.Errorf("something wrong when create data in click event : %w", err)
	}

	return clickEvent, nil
}

func (r *clickEventsRepository) FindByShortUrlId(ctx context.Context, shortUrlId int64, userId int64) ([]domain.ClickEvent, error) {
	query := `SELECT ce.id, ce.short_url_id, ce.ip_address, ce.user_agent, ce.referer, ce.clicked_at FROM click_event ce INNER JOIN short_urls su ON ce.short_url_id = su.id WHERE ce.short_url_id = $1 AND su.user_id = $2`

	rows, err := r.db.Query(ctx, query, shortUrlId, userId)
	if err != nil {
		return nil, fmt.Errorf("something wrong when find click events by short code : %w", err)
	}
	defer rows.Close()

	var clickEvents []domain.ClickEvent

	for rows.Next() {
		var clickEvent domain.ClickEvent
		err := rows.Scan(&clickEvent.Id, &clickEvent.ShortUrlId, &clickEvent.IpAddress, &clickEvent.UserAgent, &clickEvent.Referer, &clickEvent.ClickedAt)
		if err != nil {
			return nil, fmt.Errorf("something wrong when scan click event data : %w", err)
		}
		clickEvents = append(clickEvents, clickEvent)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("something wrong when iterate click events rows : %w", err)
	}

	return clickEvents, nil
}
