package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"shorter-url/internal/domain"
	"shorter-url/internal/middleware"
	"shorter-url/internal/service"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Create_ShortUrl_Pass(t *testing.T) {
	mockService := new(service.MockShortUrlService)
	mockClickService := new(service.MockClickEventService)

	handler := NewShortUrlHandler(mockService, mockClickService)

	mockService.On("CreateShortUrl", mock.Anything, int64(1), "https://www.google.com", mock.Anything).Return(&domain.ShortUrl{
		Id:          1,
		ShortCode:   "hihihi",
		OriginalUrl: "https://www.google.com",
		ExpiredAt:   time.Now().AddDate(0, 1, 0),
		CreatedAt:   time.Now(),
	}, nil)

	bodyJson := `{"original_url" : "https://www.google.com"}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "localhost:8080/api/urls", strings.NewReader(bodyJson))

	ctx := context.WithValue(request.Context(), middleware.UserIDKey, int64(1))
	request = request.WithContext(ctx)

	handler.Create(recorder, request, nil)

	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var responseBody struct {
		Message string `json:"message"`
		Data    struct {
			Id          int64     `json:"id"`
			ShortCode   string    `json:"short_code"`
			OriginalUrl string    `json:"original_url"`
			ExpiredAt   time.Time `json:"expired_at"`
			CreatedAt   time.Time `json:"created_at"`
		} `json:"data"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
	assert.NoError(t, err, "Response body harus berupa format JSON yang valid")

	assert.Equal(t, "Short code created successfully", responseBody.Message)
	assert.Equal(t, int64(1), responseBody.Data.Id)
	assert.Equal(t, "hihihi", responseBody.Data.ShortCode)
	assert.Equal(t, "https://www.google.com", responseBody.Data.OriginalUrl)

	targetExpiration := time.Now().AddDate(0, 1, 0)
	assert.WithinDuration(t, time.Now(), responseBody.Data.CreatedAt, time.Second)
	assert.WithinDuration(t, targetExpiration, responseBody.Data.ExpiredAt, time.Second)

	mockService.AssertExpectations(t)
}

func Test_Creat_ShortUrl_Unauthorized(t *testing.T) {
	mockService := new(service.MockShortUrlService)
	mockClickService := new(service.MockClickEventService)
	handler := NewShortUrlHandler(mockService, mockClickService)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "localhost:8080/api/urls", strings.NewReader(`{"original_url":"https://www.google.com"}`))

	ctx := context.WithValue(request.Context(), middleware.UserIDKey, int64(0))
	request = request.WithContext(ctx)

	handler.Create(recorder, request, nil)

	assert.Contains(t, recorder.Body.String(), "Unauthorized")
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	mockService.AssertNotCalled(t, "CreateShortUrl", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_Create_ShortUrl_InvalidJSON(t *testing.T) {
	mockService := new(service.MockShortUrlService)
	mockClickService := new(service.MockClickEventService)
	handler := NewShortUrlHandler(mockService, mockClickService)

	recorder := httptest.NewRecorder()
	brokenBodyJson := `{"original_url" : "https://www.google.com`
	request := httptest.NewRequest(http.MethodPost, "localhost:8080/api/urls", strings.NewReader(brokenBodyJson))

	ctx := context.WithValue(request.Context(), middleware.UserIDKey, int64(1))
	request = request.WithContext(ctx)

	handler.Create(recorder, request, nil)

	assert.Contains(t, recorder.Body.String(), "invalid request payload")
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	mockService.AssertNotCalled(t, "CreateShortUrl", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_Create_ShortUrl_ServiceError(t *testing.T) {
	mockService := new(service.MockShortUrlService)
	mockClickService := new(service.MockClickEventService)
	handler := NewShortUrlHandler(mockService, mockClickService)

	mockService.On("CreateShortUrl", mock.Anything, int64(1), "https://www.google.com", mock.Anything).
		Return(nil, errors.New("database connection lost"))

	bodyJson := `{"original_url" : "https://www.google.com"}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "localhost:8080/api/urls", strings.NewReader(bodyJson))

	ctx := context.WithValue(request.Context(), middleware.UserIDKey, int64(1))
	request = request.WithContext(ctx)

	handler.Create(recorder, request, nil)

	assert.Contains(t, recorder.Body.String(), "failed to create short code")
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	mockService.AssertExpectations(t)
}

func Test_AccessShortCode_Pass(t *testing.T) {
	mockService := new(service.MockShortUrlService)
	mockClickService := new(service.MockClickEventService)
	handler := NewShortUrlHandler(mockService, mockClickService)

	mockService.On("GetShortUrlByShortCode", mock.Anything, "hihihi").Return(&domain.ShortUrl{
		Id:          1,
		ShortCode:   "hihihi",
		OriginalUrl: "https://www.google.com",
	}, nil)

	mockClickService.On("Create", mock.Anything, mock.Anything).Return(&domain.ClickEvent{}, nil)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "localhost:8080/hihihi", nil)

	params := httprouter.Params{
		httprouter.Param{Key: "shortCode", Value: "hihihi"},
	}

	handler.AccessShortCode(recorder, request, params)

	assert.Equal(t, http.StatusFound, recorder.Code)
	assert.Equal(t, "https://www.google.com", recorder.Header().Get("Location"))

	mockService.AssertExpectations(t)
	mockClickService.AssertExpectations(t)
}

func Test_AccessShortCode_NotFound(t *testing.T) {
	mockService := new(service.MockShortUrlService)
	mockClickService := new(service.MockClickEventService)
	handler := NewShortUrlHandler(mockService, mockClickService)

	mockService.On("GetShortUrlByShortCode", mock.Anything, "zonk").
		Return(nil, errors.New("short code not found in database"))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "localhost:8080/zonk", nil)

	params := httprouter.Params{
		httprouter.Param{Key: "shortCode", Value: "zonk"},
	}

	handler.AccessShortCode(recorder, request, params)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.Contains(t, recorder.Body.String(), "short code not found")

	mockService.AssertExpectations(t)
	mockClickService.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func GenerateJWTToken(userId int64, email string, secretKey string) string {
	claims := jwt.MapClaims{
		"user_id": userId,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return fmt.Sprintf("failed to generated jwt token, %v", err)
	}

	return tokenString
}
