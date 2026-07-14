package service

import (
	"context"
	"fmt"
	"shorter-url/internal/domain"
	"shorter-url/internal/helper"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type userOtpsService struct {
	repo        domain.UserOtpsRepository
	mailService domain.EmailService
	userRepo    domain.UserRepository
	JwtSecret   []byte
}

func NewUserOtpsService(repo domain.UserOtpsRepository, mailService domain.EmailService, userRepo domain.UserRepository, jwtSecret []byte) domain.UserOtpsService {
	return &userOtpsService{
		repo:        repo,
		mailService: mailService,
		userRepo:    userRepo,
		JwtSecret:   jwtSecret,
	}
}

func (s *userOtpsService) SendOTP(ctx context.Context, email string, otpType string) error {
	otpCode, err := helper.GenerateRandomChar(6)
	if err != nil {
		return err
	}

	newOtps := &domain.UserOtps{
		Email:     email,
		Code:      otpCode,
		Type:      otpType,
		ExpiredAt: time.Now().Add(5 * time.Minute),
	}
	_, err = s.repo.Create(ctx, newOtps)
	if err != nil {
		return err
	}

	emailData := struct {
		Email string
		Code  string
	}{
		Email: email,
		Code:  otpCode,
	}

	var subjectMail string
	if otpType == "reset_password" {
		subjectMail = "Reset Password"
	}
	if otpType == "verification_account" {
		subjectMail = "Verification Account"
	}

	err = s.mailService.SendEmailWithHTML(ctx, email, emailData, subjectMail, otpType)
	if err != nil {
		return fmt.Errorf("send mail with html: %w", err)
	}

	return nil
}

func (s *userOtpsService) VerifyOTP(ctx context.Context, code string, email string, otpType string) (string, error) {
	result, err := s.repo.FindByEmailCodeType(ctx, email, code, otpType)
	if err != nil {
		return "", err
	}

	if result.IsUsed {
		return "", fmt.Errorf("otp code already used: %w", domain.ErrInvalidOTP)
	}
	if result.ExpiredAt.Before(time.Now()) {
		return "", fmt.Errorf("otp code expired: %w", domain.ErrInvalidOTP)
	}

	err = s.repo.UpdateIsUsed(ctx, result.Id, true)
	if err != nil {
		return "", err
	}

	if otpType == "verification_account" {
		user, err := s.userRepo.FindByEmail(ctx, result.Email)
		if err != nil {
			return "", err
		}
		err = s.userRepo.UpdateVerified(ctx, user.Id, true)
		if err != nil {
			return "", err
		}
		err = s.repo.DeleteByCodeOTP(ctx, result.Code)
		if err != nil {
			return "", err
		}
		return "", nil
	}

	claims := jwt.MapClaims{
		"email":    email,
		"otp_type": otpType,
		"exp":      time.Now().Add(5 * time.Minute).Unix(),
	}

	token, err := helper.GenerateJWTToken(claims, s.JwtSecret)
	if err != nil {
		return "", err
	}

	return token, nil
}

