package domain

import "errors"

var (
	ErrNotFound               = errors.New("resource not found")
	ErrNoRows                 = errors.New("no rows affected")
	ErrInvalidOTP             = errors.New("invalid code otp")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrInvalidToken           = errors.New("invalid token")
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrInvalidOauthRequest    = errors.New("invalid oauth request")
)
