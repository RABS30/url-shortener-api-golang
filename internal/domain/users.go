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

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id int64) error
	UpdatePassword(ctx context.Context, id int64, hashedPassword string) error
	UpdateVerified(ctx context.Context, id int64, verify bool) error
	FindById(ctx context.Context, id int64) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
}

type UserService interface {
	Register(ctx context.Context, email string, password string) (*User, error)
	Login(ctx context.Context, email string, password string) (string, error)
	ChangePassword(ctx context.Context, email string, oldPassword string, newPassword string) error
	ResetPassword(ctx context.Context, newPassword string, resetToken string) error
}

type EmailService interface {
	SendEmail(ctx context.Context, to string, subject string, body string) error
	SendEmailWithHTML(ctx context.Context, to string, context any, subjectText string, templateName string) error
}

type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Compare(ctx context.Context, password string, hashedPassword string) error
}
