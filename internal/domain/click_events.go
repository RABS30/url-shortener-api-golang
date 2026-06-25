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
	ClickedAt  time.Time
}

type ClickEventRepository interface {
	Create(ctx context.Context, clickEvent *ClickEvent) (*ClickEvent, error)
	FindByShortUrlId(ctx context.Context, shortUrlId int64, userId int64) ([]ClickEvent, error)
}

type ClickEventService interface {
	Create(ctx context.Context, clickEvent *ClickEvent) (*ClickEvent, error)
	FindByShortUrlId(ctx context.Context, shortUrlId int64, userId int64) ([]ClickEvent, error)
}
