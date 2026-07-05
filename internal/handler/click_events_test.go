package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"shorter-url/internal/domain"
	"shorter-url/internal/middleware"
	"shorter-url/internal/service"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_FindByShortUrlId_Pass(t *testing.T) {
	mockService := new(service.MockClickEventService)
	h := NewClickEventHandler(mockService)

	claims := &middleware.UserPrimaryClaims{
		UserID: 1,
		Email:  "test@gmail.com",
	}
	req := httptest.NewRequest(http.MethodGet, "/click-events/42", nil)
	ctx := context.WithValue(req.Context(), middleware.UserClaimsKey, claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	expectedList := []domain.ClickEvent{
		{Id: 10, ShortUrlId: 42, IpAddress: "1.1.1.1"},
	}

	mockService.On("FindByShortUrlId", mock.Anything, int64(42), int64(1)).Return(expectedList, nil)

	h.FindByShortUrlId(rec, req, httprouter.Params{
		httprouter.Param{Key: "shortUrlId", Value: "42"},
	})

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ok")
	mockService.AssertExpectations(t)
}

func Test_FindByShortUrlId_Fail_Unauthorized(t *testing.T) {
	mockService := new(service.MockClickEventService)
	h := NewClickEventHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/click-events/42", nil)
	rec := httptest.NewRecorder()

	h.FindByShortUrlId(rec, req, httprouter.Params{
		httprouter.Param{Key: "shortUrlId", Value: "42"},
	})

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "unauthorized")
}

func Test_FindByShortUrlId_Fail_InvalidShortUrlId(t *testing.T) {
	mockService := new(service.MockClickEventService)
	h := NewClickEventHandler(mockService)

	claims := &middleware.UserPrimaryClaims{
		UserID: 1,
		Email:  "test@gmail.com",
	}
	req := httptest.NewRequest(http.MethodGet, "/click-events/42", nil)
	ctx := context.WithValue(req.Context(), middleware.UserClaimsKey, claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	h.FindByShortUrlId(rec, req, httprouter.Params{
		httprouter.Param{Key: "shortUrlId", Value: "invalid"},
	})

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid short url id")
}

func Test_FindByShortUrlId_Fail_NotFound(t *testing.T) {
	mockService := new(service.MockClickEventService)
	h := NewClickEventHandler(mockService)

	claims := &middleware.UserPrimaryClaims{
		UserID: 1,
		Email:  "test@gmail.com",
	}
	req := httptest.NewRequest(http.MethodGet, "/click-events/42", nil)
	ctx := context.WithValue(req.Context(), middleware.UserClaimsKey, claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	mockService.On("FindByShortUrlId", mock.Anything, int64(42), int64(1)).Return(nil, errors.New("not found"))

	h.FindByShortUrlId(rec, req, httprouter.Params{
		httprouter.Param{Key: "shortUrlId", Value: "42"},
	})

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "short url data not found")
	mockService.AssertExpectations(t)
}
