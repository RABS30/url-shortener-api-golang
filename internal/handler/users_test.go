package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"shorter-url/internal/domain"
	"shorter-url/internal/service"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =========================================================================
// UNIT TEST FOR REGISTER
// =========================================================================

func Test_Register_Pass(t *testing.T) {
	mockAuth := new(service.MockAuthService)
	handler := NewUserHandler(mockAuth)

	mockAuth.On("Register", mock.Anything, "test@mail.com", "password123").Return(&domain.User{
		Id:    1,
		Email: "test@mail.com",
	}, nil)

	body := `{"email":"test@mail.com", "password":"password123"}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(body))

	handler.Register(recorder, request, nil)

	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "registration successful")
	mockAuth.AssertExpectations(t)
}

func Test_Register_InvalidJSON(t *testing.T) {
	mockAuth := new(service.MockAuthService)
	handler := NewUserHandler(mockAuth)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(`{"email":`))

	handler.Register(recorder, request, nil)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "invalid json format")
	mockAuth.AssertNotCalled(t, "Register", mock.Anything, mock.Anything, mock.Anything)
}

func Test_Register_RequiredFields(t *testing.T) {
	mockAuth := new(service.MockAuthService)
	handler := NewUserHandler(mockAuth)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(`{"email":"", "password":""}`))

	handler.Register(recorder, request, nil)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "email and password are required")
	mockAuth.AssertNotCalled(t, "Register", mock.Anything, mock.Anything, mock.Anything)
}

func Test_Register_ServiceError(t *testing.T) {
	mockAuth := new(service.MockAuthService)
	handler := NewUserHandler(mockAuth)

	mockAuth.On("Register", mock.Anything, "duplicate@mail.com", "password123").Return(nil, errors.New("email already exists"))

	body := `{"email":"duplicate@mail.com", "password":"password123"}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(body))

	handler.Register(recorder, request, nil)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "email already exists")
	mockAuth.AssertExpectations(t)
}


func Test_Login_Pass(t *testing.T) {
	mockAuth := new(service.MockAuthService)
	handler := NewUserHandler(mockAuth)

	mockAuth.On("Login", mock.Anything, "test@mail.com", "password123").Return("mocked-jwt-token", nil)

	body := `{"email":"test@mail.com", "password":"password123"}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))

	handler.Login(recorder, request, nil)

	// 1. Validasi HTTP Status
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "login successfully")

	// 2. Validasi Cookie (Penting untuk skenario Login Anda!)
	cookies := recorder.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "token", cookies[0].Name)
	assert.Equal(t, "mocked-jwt-token", cookies[0].Value)
	assert.True(t, cookies[0].HttpOnly)

	mockAuth.AssertExpectations(t)
}

func Test_Login_InvalidJSON(t *testing.T) {
	mockAuth := new(service.MockAuthService)
	handler := NewUserHandler(mockAuth)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"email":`))

	handler.Login(recorder, request, nil)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "invalid json format")
	mockAuth.AssertNotCalled(t, "Login", mock.Anything, mock.Anything, mock.Anything)
}

func Test_Login_RequiredFields(t *testing.T) {
	mockAuth := new(service.MockAuthService)
	handler := NewUserHandler(mockAuth)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"email":"test@mail.com", "password":""}`))

	handler.Login(recorder, request, nil)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "email and password are required")
	mockAuth.AssertNotCalled(t, "Login", mock.Anything, mock.Anything, mock.Anything)
}

func Test_Login_Unauthorized(t *testing.T) {
	mockAuth := new(service.MockAuthService)
	handler := NewUserHandler(mockAuth)

	mockAuth.On("Login", mock.Anything, "wrong@mail.com", "badpass").Return("", errors.New("invalid email or password"))

	body := `{"email":"wrong@mail.com", "password":"badpass"}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))

	handler.Login(recorder, request, nil)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "invalid email or password")
	mockAuth.AssertExpectations(t)
}
