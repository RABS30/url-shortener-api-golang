package domain

import (
	"context"
	"time"
)

type ClickEvent struct {
	Id         int64
	ShortUrlId int64
	IpAddress  string
	UserAgent  string
	Referer    string
	ClickedAt  string
}

type ClickEventRepository interface {
	Create(ctx context.Context, passwordResetToken *ClickEvent) (*ClickEvent, error)
	Delete(ctx context.Context, id int64) error
	FindById(ctx context.Context, id int64) (*ClickEvent, error)
	FindByShortCode(ctx context.Context, shortCode string) ([]ClickEvent, error)
	FilterByDate(ctx context.Context, date time.Time) ([]ClickEvent, error)
	FindAll(ctx context.Context) ([]ClickEvent, error)
}
