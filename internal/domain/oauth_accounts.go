package domain

import (
	"context"
	"time"
)

type OauthAccounts struct {
	Id             int64     `json:"id"`
	UserId         int64     `json:"user_id"`
	Provider       string    `json:"provider"`
	ProviderUserId string    `json:"provider_user_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type OauthAccountsRepository interface {
	Upsert(ctx context.Context, oauthAccount *OauthAccounts) (*OauthAccounts, error)
	FindById(ctx context.Context, id int64) (*OauthAccounts, error)
	FindByProviderUserId(ctx context.Context, provider, providerUserId string) (*OauthAccounts, error)
	FindByUserIdAndProvider(ctx context.Context, userId int64, provider string) (*OauthAccounts, error)
	Update(ctx context.Context, oauthAccounts *OauthAccounts) (*OauthAccounts, error)
	DeleteById(ctx context.Context, id int64) error
}
