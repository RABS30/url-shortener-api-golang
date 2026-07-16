package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"shorter-url/internal/database"
	"shorter-url/internal/domain"
	"shorter-url/internal/helper"
	"shorter-url/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

type userService struct {
	Repo       domain.UserRepository
	JwtSecret  []byte
	Hasher     domain.PasswordHasher
	Db         database.PgxTransactor
	OtpService domain.UserOtpsService
}

func NewUserService(repo domain.UserRepository, JwtSecret []byte, hasher domain.PasswordHasher, db database.PgxTransactor, otpsService domain.UserOtpsService) domain.UserService {
	return &userService{
		Repo:       repo,
		JwtSecret:  JwtSecret,
		Hasher:     hasher,
		Db:         db,
		OtpService: otpsService,
	}
}

func (s *userService) Register(ctx context.Context, email string, password string) (*domain.User, error) {
	existingUser, err := s.Repo.FindByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("failed to check existing email: %w", err)
		}
	}
	if existingUser != nil {
		return nil, domain.ErrEmailAlreadyRegistered
	}

	hashedPassword, err := s.Hasher.Hash(ctx, password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	result, err := s.Repo.Create(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	otpCtx := context.WithoutCancel(ctx)
	go func() {
		err = s.OtpService.SendOTP(otpCtx, result.Email, "verification_account")
		if err != nil {
			log.Printf("failed to  send otp code: %v", err)
		}
	}()

	return result, nil
}

func (s *userService) Login(ctx context.Context, email string, password string) (string, error) {

	existingUser, err := s.Repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrInvalidEmailorPassword
		}
		return "", fmt.Errorf("failed to find user by email: %w", err)
	}
	if existingUser == nil {
		return "", domain.ErrInvalidEmailorPassword
	}

	if !existingUser.IsVerified {
		return "", domain.ErrUnverified
	}

	err = s.Hasher.Compare(ctx, password, existingUser.PasswordHash)
	if err != nil {
		return "", domain.ErrInvalidEmailorPassword
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
	user, err := s.Repo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	err = s.Hasher.Compare(ctx, oldPassword, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("compare password in change password: %w", domain.ErrInvalidCredentials)
	}

	hashedPassword, err := s.Hasher.Hash(ctx, newPassword)
	if err != nil {
		return fmt.Errorf("new hash password in change password: %w", err)
	}

	err = s.Repo.UpdatePassword(ctx, user.Id, string(hashedPassword))
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

	hashedPassword, err := s.Hasher.Hash(ctx, newPassword)
	if err != nil {
		return fmt.Errorf("create hash password in reset password: %w", err)
	}

	user, err := s.Repo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	err = s.Repo.UpdatePassword(ctx, user.Id, string(hashedPassword))
	if err != nil {
		return err
	}

	return nil
}

func (s *userService) LoginWithGoogle(ctx context.Context, info *domain.GoogleUserInfo) (string, error) {
	tx, err := s.Db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("login with google: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	txUserRepo := repository.NewUserRepository(tx)
	txOauthRepo := repository.NewOauthAccountsRepository(tx)

	fmt.Println()
	fmt.Println(info)
	fmt.Println()

	user, err := txUserRepo.Upsert(ctx, &domain.User{
		Email:      info.Email,
		IsVerified: info.EmailVerified,
	})
	if err != nil {
		return "", fmt.Errorf("login with google: upsert user: %w", err)
	}

	_, err = txOauthRepo.Upsert(ctx, &domain.OauthAccounts{
		UserId:         user.Id,
		Provider:       "google",
		ProviderUserId: info.GoogleID,
	})
	if err != nil {
		return "", fmt.Errorf("login with google: upsert oauth account: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("login with google: commit tx: %w", err)
	}

	claims := jwt.MapClaims{
		"user_id": user.Id,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token, err := helper.GenerateJWTToken(claims, s.JwtSecret)
	if err != nil {
		return "", fmt.Errorf("login with google: generate jwt: %w", err)
	}
	return token, nil
}
