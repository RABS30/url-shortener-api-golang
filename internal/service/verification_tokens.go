package service

import (
	"context"
	"errors"
	"fmt"
	"shorter-url/internal/domain"
	"shorter-url/internal/helper"
	"time"
)

type verificationTokenService struct {
	repo         domain.VerificationTokenRepository
	users        domain.UserRepository
	emailService domain.EmailService
	baseUrl      string
}

func NewVerificationTokenService(repo domain.VerificationTokenRepository, users domain.UserRepository, emailService domain.EmailService, baseUrl string) domain.VerificationTokenService {
	return &verificationTokenService{
		repo:         repo,
		users:        users,
		emailService: emailService,
		baseUrl:      baseUrl,
	}
}

func (s *verificationTokenService) SendVerificationMail(ctx context.Context, email string) error {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user.IsVerified {
		return errors.New("user is verified")
	}

	token, err := helper.GenerateRandomToken(16)
	if err != nil {
		return err
	}

	context := struct {
		VerificationURL string
	}{
		VerificationURL: fmt.Sprintf("%s/verify?token=%s", s.baseUrl, token),
	}

	verificationTokenData := &domain.VerificationToken{
		UserId:    user.Id,
		Token:     token,
		ExpiredAt: time.Now().Add(time.Minute * 15),
	}

	_, err = s.repo.Create(ctx, verificationTokenData)
	if err != nil {
		return err
	}

	err = s.emailService.SendEmailWithHTML(ctx, user.Email, context, "verification_account_mail.html")
	if err != nil {
		return err
	}

	return nil
}

func (s *verificationTokenService) VerificationAccount(ctx context.Context, token string) error {
	dataToken, err := s.repo.FindByToken(ctx, token)
	if err != nil {
		return err
	}
	if time.Now().After(dataToken.ExpiredAt) {
		return errors.New("token has expired")
	}

	user, err := s.users.FindById(ctx, dataToken.UserId)
	if err != nil {
		return err
	}

	if user.IsVerified {
		return errors.New("user is already verified")
	}

	err = s.users.UpdateVerified(ctx, dataToken.UserId, true)
	if err != nil {
		return err
	}

	err = s.repo.DeleteByUserId(ctx, dataToken.UserId)
	if err != nil {
		return err
	}

	return nil
}
