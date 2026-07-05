package domain

import "errors"

var (
	ErrNotFound        = errors.New("resource not found")
	ErrNoRows          = errors.New("no rows affected")
	ErrInvalidOTP      = errors.New("invalid code otp")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidToken    = errors.New("invalid token")
)
