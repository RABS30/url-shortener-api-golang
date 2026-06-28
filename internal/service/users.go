package service

import (
	"context"
	"errors"
	"fmt"
	"shorter-url/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo      domain.UserRepository
	JwtSecret []byte
}

func NewUserService(repo domain.UserRepository, JwtSecret []byte) domain.AuthService {
	return &userService{
		repo:      repo,
		JwtSecret: JwtSecret,
	}
}

func (s *userService) Register(ctx context.Context, email string, password string) (*domain.User, error) {
	existingUser, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("something error with database,  %w", err)
		}
	}
	if existingUser != nil {
		return nil, errors.New("email already exist, use another email")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password, %w", err)
	}

	newUser := &domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	result, err := s.repo.Create(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create new account, %w", err)
	}

	return result, nil
}

func (s *userService) Login(ctx context.Context, email string, password string) (string, error) {
	invalidError := errors.New("Invalid email and password")

	existingUser, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("cannot find email, %w", err)
	}
	if existingUser == nil {
		return "", invalidError
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingUser.PasswordHash), []byte(password))
	if err != nil {
		return "", invalidError
	}

	claims := jwt.MapClaims{
		"user_id":     existingUser.Id,
		"email":       existingUser.Email,
		"is_verified": existingUser.IsVerified,
		"exp":         time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.JwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generated jwt token, %w", err)
	}

	return tokenString, nil
}
