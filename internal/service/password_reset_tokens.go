package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"shorter-url/internal/domain"
	"shorter-url/internal/helper"
	"time"
)

type passwordResetTokensService struct {
	repo         domain.PasswordResetTokensRepository
	users        domain.UserRepository
	emailService domain.EmailService
	hasher       domain.PasswordHasher
	baseUrl      string
}

func NewPasswordResetTokensService(repo domain.PasswordResetTokensRepository, users domain.UserRepository, emailService domain.EmailService, hasher domain.PasswordHasher, baseUrl string) domain.PasswordResetTokensService {
	return &passwordResetTokensService{
		repo:         repo,
		users:        users,
		emailService: emailService,
		hasher:       hasher,
		baseUrl:      baseUrl,
	}
}

func (s *passwordResetTokensService) RequestResetPassword(ctx context.Context, email string) error {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	token, err := helper.GenerateRandomToken(16)
	if err != nil {
		return err
	}

	var tokenUser = &domain.PasswordResetTokens{
		UserId:    user.Id,
		Token:     token,
		ExpiredAt: time.Now().Add(time.Minute * 15),
	}

	_, err = s.repo.Create(ctx, tokenUser)
	if err != nil {
		return err
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.baseUrl, token)

	subject := "Permintaan Reset Password Akun Anda"
	body := fmt.Sprintf(`
		<h3>Halo, %s</h3>
		<p>Kami menerima permintaan untuk mereset password akun Anda.</p>
		<p>Silakan klik link di bawah ini untuk melanjutkan proses reset password:</p>
		<p><a href="%s" style="padding: 10px 20px; background-color: #007bff; color: white; text-decoration: none; border-radius: 5px;">Reset Password Saya</a></p>
		<br>
		<p><i>Link ini hanya berlaku selama 15 menit. Jika Anda tidak merasa melakukan permintaan ini, abaikan email ini.</i></p>
	`, user.Email, resetURL)

	err = s.emailService.SendEmail(ctx, user.Email, subject, body)
	if err != nil {
		return fmt.Errorf("failed to send email, %w", err)
	}

	return nil
}

func (s *passwordResetTokensService) ExecuteResetPassword(ctx context.Context, token string, password1 string, password2 string) error {
	userToken, err := s.repo.FindByToken(ctx, token)
	if err != nil {
		return err
	}
	if time.Now().After(userToken.ExpiredAt) {
		return errors.New("token expired")
	}
	if password1 != password2 {
		return errors.New("password not equal")
	}

	hashedPassword, err := s.hasher.Hash(ctx, password1)
	if err != nil {
		return err
	}

	err = s.users.UpdatePassword(ctx, userToken.UserId, hashedPassword)
	if err != nil {
		return err
	}

	err = s.repo.DeleteByUserId(ctx, userToken.UserId)
	if err != nil {
		log.Printf("[WARNING] failed to delete password reset token for userId %d: %v", userToken.UserId, err)
	}

	return nil
}
