package repository

import (
	"context"
	"fmt"
	"shorter-url/internal/domain"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type clickEventsRepository struct {
	db *pgxpool.Pool
}

func NewClickEventsRepository(db *pgxpool.Pool) domain.ClickEventRepository {
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

func (r *clickEventsRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM click_events WHERE id = $1"

	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("something wrong when delete  data in click event : %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("there is no data deleted, click event with ID %d not found", id)
	}

	return nil
}

func (r *clickEventsRepository) FindById(ctx context.Context, id int64) (*domain.ClickEvent, error) {
	query := `SELECT id, ip_address, short_url_id, user_agent, referer, clicked_at FROM click_events WHERE id = $1`

	var clickEvent domain.ClickEvent
	err := r.db.QueryRow(ctx, query, id).Scan(&clickEvent.Id, &clickEvent.IpAddress, &clickEvent.ShortUrlId, &clickEvent.UserAgent, &clickEvent.Referer, &clickEvent.ClickedAt)
	if err != nil {
		return nil, fmt.Errorf("something wrong when find click event by ID : %w", err)
	}

	return &clickEvent, nil
}

func (r *clickEventsRepository) FindByShortCode(ctx context.Context, shortCode string) ([]domain.ClickEvent, error) {
	query := `SELECT ce.id, ce.ip_address, ce.short_url_id, ce.user_agent, ce.referer, ce.clicked_at
	FROM click_events ce
	JOIN short_urls su ON ce.short_url_id = su.id
	WHERE su.short_code = $1`

	rows, err := r.db.Query(ctx, query, shortCode)
	if err != nil {
		return nil, fmt.Errorf("something wrong when find click events by short code : %w", err)
	}

	defer rows.Close()

	var clickEvents []domain.ClickEvent

	for rows.Next() {
		var clickEvent domain.ClickEvent
		err := rows.Scan(&clickEvent.Id, &clickEvent.IpAddress, &clickEvent.ShortUrlId, &clickEvent.UserAgent, &clickEvent.Referer, &clickEvent.ClickedAt)
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

func (r *clickEventsRepository) FilterByDate(ctx context.Context, date time.Time) ([]domain.ClickEvent, error) {
	query := `SELECT id, ip_address, short_url_id, user_agent, referer, clicked_at FROM click_events WHERE DATE(clicked_at) = $1`

	rows, err := r.db.Query(ctx, query, date.Format("2006-01-02"))

	if err != nil {
		return nil, fmt.Errorf("something wrong when filter click events by date : %w", err)
	}

	defer rows.Close()

	var clickEvents []domain.ClickEvent

	for rows.Next() {
		var clickEvent domain.ClickEvent
		err := rows.Scan(&clickEvent.Id, &clickEvent.IpAddress, &clickEvent.ShortUrlId, &clickEvent.UserAgent, &clickEvent.Referer, &clickEvent.ClickedAt)
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

func (r *clickEventsRepository) FindAll(ctx context.Context) ([]domain.ClickEvent, error) {
	query := `SELECT id, ip_address, short_url_id, user_agent, referer, clicked_at FROM click_events`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("something error when get all data : %w", err)
	}

	defer rows.Close()

	var clickEvents []domain.ClickEvent

	for rows.Next() {
		var clickEvent domain.ClickEvent
		err := rows.Scan(&clickEvent.Id, &clickEvent.IpAddress, &clickEvent.ShortUrlId, &clickEvent.UserAgent, &clickEvent.Referer, &clickEvent.ClickedAt)
		if err != nil {
			return nil, fmt.Errorf("something error when scan click event data : %w", err)
		}

		clickEvents = append(clickEvents, clickEvent)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("something error when iterate click events rows : %w", err)
	}

	return clickEvents, nil
}
