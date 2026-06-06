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
	FindById(ctx context.Context, id int64) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
}
