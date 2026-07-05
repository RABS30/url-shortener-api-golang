package service

import (
	"context"
	"errors"
	"fmt"
	"shorter-url/internal/domain"
	"shorter-url/internal/helper"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

type userService struct {
	repo      domain.UserRepository
	JwtSecret []byte
	hasher    domain.PasswordHasher
}

func NewUserService(repo domain.UserRepository, JwtSecret []byte, hasher domain.PasswordHasher) domain.UserService {
	return &userService{
		repo:      repo,
		JwtSecret: JwtSecret,
		hasher:    hasher,
	}
}

func (s *userService) Register(ctx context.Context, email string, password string) (*domain.User, error) {
	existingUser, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing email: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := s.hasher.Hash(ctx, password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	result, err := s.repo.Create(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return result, nil
}

func (s *userService) Login(ctx context.Context, email string, password string) (string, error) {
	invalidError := errors.New("invalid email or password")

	existingUser, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", invalidError
		}
		return "", fmt.Errorf("failed to find user by email: %w", err)
	}
	if existingUser == nil {
		return "", invalidError
	}

	err = s.hasher.Compare(ctx, password, existingUser.PasswordHash)
	if err != nil {
		return "", invalidError
	}

	claims := jwt.MapClaims{
		"user_id": existingUser.Id,
		"email":   existingUser.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	tokenString, err := helper.GenerateJWTToken(claims, s.JwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *userService) ChangePassword(ctx context.Context, email string, oldPassword string, newPassword string) error {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	err = s.hasher.Compare(ctx, oldPassword, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("compare password in change password: %w", domain.ErrInvalidPassword)
	}

	hashedPassword, err := s.hasher.Hash(ctx, newPassword)
	if err != nil {
		return fmt.Errorf("new hash password in change password: %w", err)
	}

	err = s.repo.UpdatePassword(ctx, user.Id, string(hashedPassword))
	if err != nil {
		return err
	}
	return nil
}

func (s *userService) ResetPassword(ctx context.Context, newPassword string, resetToken string) error {
	token, err := jwt.Parse(resetToken, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.JwtSecret, nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("reset password: %w", domain.ErrInvalidToken)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["otp_type"] != "reset_password" {
		return fmt.Errorf("invalid token claims: %w", domain.ErrInvalidToken)
	}

	email, _ := claims["email"].(string)

	hashedPassword, err := s.hasher.Hash(ctx, newPassword)
	if err != nil {
		return fmt.Errorf("create hash password in reset password: %w", err)
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	err = s.repo.UpdatePassword(ctx, user.Id, string(hashedPassword))
	if err != nil {
		return err
	}

	return nil
}
