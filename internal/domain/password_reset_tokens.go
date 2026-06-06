package domain

import (
	"context"
	"time"
)

type PasswordResetTokens struct {
	Id        int64
	UserId    int64
	Token     string
	ExpiredAt time.Time
}

type PasswordResetTokensRepository interface {
	Create(ctx context.Context, passwordResetToken *PasswordResetTokens) (*PasswordResetTokens, error)
	Delete(ctx context.Context, id int64) error
	FindByToken(ctx context.Context, token string) (*PasswordResetTokens, error)
}
