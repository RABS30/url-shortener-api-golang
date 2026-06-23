package service

import (
	"context"
	"errors"
	"shorter-url/internal/domain"
	"shorter-url/internal/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_ShortUrlService_Create_Pass(t *testing.T) {
	mockRepo := new(repository.MockShortUrlsRepository)
	svc := NewShortUrlService(mockRepo)

	expectedResult := &domain.ShortUrl{Id: 1, OriginalUrl: "https://rabs.dev"}
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(expectedResult, nil)

	result, err := svc.CreateShortUrl(context.Background(), int64(1), "https://rabs.dev", time.Now())

	assert.NoError(t, err)
	assert.Equal(t, "https://rabs.dev", result.OriginalUrl)
	mockRepo.AssertExpectations(t)
}

func Test_ShortUrlService_Create_InvalidURL(t *testing.T) {
	mockRepo := new(repository.MockShortUrlsRepository)
	svc := NewShortUrlService(mockRepo)

	// Mengirim URL cacat (tanpa http/https scheme)
	result, err := svc.CreateShortUrl(context.Background(), int64(1), "bukan-url-valid", time.Now())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "URL format not valid")

	// Database tidak boleh disentuh sama sekali jika format URL sudah cacat
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func Test_ShortUrlService_Create_RetryMaxThreeTimes(t *testing.T) {
	mockRepo := new(repository.MockShortUrlsRepository)
	svc := NewShortUrlService(mockRepo)

	// Memicu simulasi error database (misal short code bentrok terus-menerus)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("duplicate key error"))

	result, err := svc.CreateShortUrl(context.Background(), int64(1), "https://google.com", time.Now())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to create short URL")

	// SANGAT PENTING: Memastikan database benar-benar dicoba dipanggil sebanyak 3 kali (Looping)
	mockRepo.AssertNumberOfCalls(t, "Create", 3)
}

func Test_ShortUrlService_Delete_Pass(t *testing.T) {
	mockRepo := new(repository.MockShortUrlsRepository)
	svc := NewShortUrlService(mockRepo)

	mockRepo.On("Delete", mock.Anything, int64(10)).Return(nil)

	err := svc.DeleteShortUrl(context.Background(), int64(10))

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func Test_ShortUrlService_GetByShortCode_NotFound(t *testing.T) {
	mockRepo := new(repository.MockShortUrlsRepository)
	svc := NewShortUrlService(mockRepo)

	mockRepo.On("FindByShortCode", mock.Anything, "missing").Return(nil, errors.New("sql: no rows in result set"))

	result, err := svc.GetShortUrlByShortCode(context.Background(), "missing")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get data by short code")
	mockRepo.AssertExpectations(t)
}

func Test_ShortUrlService_GetByUserId_Pass(t *testing.T) {
	mockRepo := new(repository.MockShortUrlsRepository)
	svc := NewShortUrlService(mockRepo)

	mockData := []domain.ShortUrl{
		{Id: 1, ShortCode: "AzdW2a"},
		{Id: 2, ShortCode: "Ab6Fxc"},
	}
	mockRepo.On("FindByUserId", mock.Anything, int64(1)).Return(mockData, nil)

	result, err := svc.GetShortUrlsByUserId(context.Background(), int64(1))

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "AzdW2a", result[0].ShortCode)
	mockRepo.AssertExpectations(t)
}
