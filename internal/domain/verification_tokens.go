package domain

import (
	"context"
	"time"
)

type VerificationToken struct {
	Id        int64     `json:"id"`
	UserId    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}

type VerificationTokenRepository interface {
	Create(ctx context.Context, verificationToken *VerificationToken) (*VerificationToken, error)
	Delete(ctx context.Context, id int64) error
	FindByToken(ctx context.Context, token string) (*VerificationToken, error)
	DeleteByUserId(ctx context.Context, userId int64) error
}

type VerificationTokenService interface {
	SendVerificationMail(ctx context.Context, email string) error
	VerificationAccount(ctx context.Context, token string) error
}
