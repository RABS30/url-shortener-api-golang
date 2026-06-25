package service

import (
	"context"
	"errors"
	"shorter-url/internal/domain"
	"shorter-url/internal/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Create_Pass(t *testing.T) {
	mockRepo := new(repository.MockClickEventRepository)
	ctx := context.Background()
	dateNow := time.Now()

	input := &domain.ClickEvent{ShortUrlId: 1, IpAddress: "127.0.0.1"}
	expectedResult := &domain.ClickEvent{Id: 10, ShortUrlId: 1, IpAddress: "127.0.0.1", ClickedAt: dateNow}

	mockRepo.On("Create", ctx, input).Return(expectedResult, nil)

	service := NewClickEventService(mockRepo)
	result, err := service.Create(ctx, input)

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	mockRepo.AssertExpectations(t)
}

func Test_Create_Fail(t *testing.T) {
	mockRepo := new(repository.MockClickEventRepository)
	ctx := context.Background()

	input := &domain.ClickEvent{ShortUrlId: 1}
	mockRepo.On("Create", ctx, input).Return(nil, errors.New("db error"))

	service := NewClickEventService(mockRepo)
	result, err := service.Create(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "something error when create click event")
	mockRepo.AssertExpectations(t)
}

func Test_FindByShortUrlId_Pass(t *testing.T) {
	mockRepo := new(repository.MockClickEventRepository)
	ctx := context.Background()
	var shortUrlId int64 = 1
	var userId int64 = 1

	expectedList := []domain.ClickEvent{
		{Id: 1, ShortUrlId: shortUrlId, IpAddress: "192.168.1.1"},
		{Id: 2, ShortUrlId: shortUrlId, IpAddress: "192.168.1.2"},
	}

	// KUNCI PERBAIKAN: Tambahkan variabel 'userId' di sini rabs!
	mockRepo.On("FindByShortUrlId", ctx, shortUrlId, userId).Return(expectedList, nil)

	service := NewClickEventService(mockRepo)
	result, err := service.FindByShortUrlId(ctx, shortUrlId, userId)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "192.168.1.1", result[0].IpAddress)
	mockRepo.AssertExpectations(t)
}

func Test_FindByShortUrlId_Fail(t *testing.T) {
	mockRepo := new(repository.MockClickEventRepository)
	ctx := context.Background()
	var shortUrlId int64 = 1
	var userId int64 = 1

	// KUNCI PERBAIKAN: Tambahkan variabel 'userId' di sini juga!
	mockRepo.On("FindByShortUrlId", ctx, shortUrlId, userId).Return(nil, errors.New("query timeout"))

	service := NewClickEventService(mockRepo)
	result, err := service.FindByShortUrlId(ctx, shortUrlId, userId)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "something error when get list click event")
	mockRepo.AssertExpectations(t)
}