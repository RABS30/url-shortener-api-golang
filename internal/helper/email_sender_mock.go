package helper

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockEmailSender struct {
	mock.Mock
}

func (m *MockEmailSender) SendEmail(ctx context.Context, to, subject, body string) error {
	return m.Called(ctx, to, subject, body).Error(0)
}
