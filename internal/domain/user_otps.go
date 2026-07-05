package domain

import (
	"context"
	"time"
)

type UserOtps struct {
	Id        int64     `json:"id"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	Type      string    `json:"type"`
	IsUsed    bool      `json:"is_used"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}

type UserOtpsRepository interface {
	Create(ctx context.Context, userOtps *UserOtps) (*UserOtps, error)
	Delete(ctx context.Context, id int64) error
	DeleteByCodeOTP(ctx context.Context, code string) error
	UpdateIsUsed(ctx context.Context, id int64, isUsed bool) error
	FindById(ctx context.Context, id int64) (*UserOtps, error)
	FindByEmailCodeType(ctx context.Context, email string, code string, otpType string) (*UserOtps, error)
}

type UserOtpsService interface {
	SendOTP(ctx context.Context, email string, otpType string) error
	VerifyOTP(ctx context.Context, code string, email string, otpType string) (string, error)
}
