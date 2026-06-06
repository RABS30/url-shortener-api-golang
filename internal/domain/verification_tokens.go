package domain

import (
	"context"
	"time"
)

type VerificationToken struct {
	Id        int64
	UserId    int64
	Token     string
	ExpiredAt time.Time
}

type VerificationTokenRepository interface {
	Create(ctx context.Context, verificationToken *VerificationToken) (*VerificationToken, error)
	Delete(ctx context.Context, id int64) error
	FindByToken(ctx context.Context, token string) (*VerificationToken, error)
}
