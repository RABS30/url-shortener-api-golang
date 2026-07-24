package domain

import (
	"context"
	"time"
)

type User struct {
	Id           int64
	Email        string
	PasswordHash string
	IsVerified   bool
	Status       string
	CreatedAt    time.Time
}

type GoogleUserInfo struct {
	GoogleID      string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

type VerificationAccountContext struct {
	Email string `json:"email"`
	Code  string
}

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id int64) error
	UpdatePassword(ctx context.Context, id int64, hashedPassword string) error
	UpdateVerified(ctx context.Context, id int64, verify bool) error
	FindById(ctx context.Context, id int64) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Upsert(ctx context.Context, user *User) (*User, error)
}

type UserService interface {
	Register(ctx context.Context, email string, password string) (*User, error)
	Login(ctx context.Context, email string, password string) (string, error)
	ChangePassword(ctx context.Context, userId int64, oldPassword string, newPassword string) error
	ResetPassword(ctx context.Context, newPassword string, resetToken string) error
	LoginWithGoogle(ctx context.Context, userInfo *GoogleUserInfo) (string, error)
}

type EmailService interface {
	SendEmail(ctx context.Context, to string, subject string, body string) error
	SendEmailWithHTML(ctx context.Context, to string, context any, subjectText string, templateName string) error
}

type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Compare(ctx context.Context, password string, hashedPassword string) error
}
