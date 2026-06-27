package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"shorter-url/internal/middleware"
	"shorter-url/internal/service"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_ForgotPasswordHandler_Pass(t *testing.T) {
	mockService := new(service.MockPasswordResetTokensService)

	jsonPayload := `{"email": "rabs@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewBufferString(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mockService.On("RequestResetPassword", mock.Anything, "rabs@example.com").Return(nil)

	h := NewPasswordResetTokensHandler(mockService)
	h.ForgotPasswordHandler(rec, req, httprouter.Params{})

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "success")
	mockService.AssertExpectations(t)
}

func Test_ForgotPasswordHandler_Fail_InvalidPayload(t *testing.T) {
	mockService := new(service.MockPasswordResetTokensService)

	jsonPayload := `{"email": "rabs@example.com", "unknown_field": "error"}`
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewBufferString(jsonPayload))
	rec := httptest.NewRecorder()

	h := NewPasswordResetTokensHandler(mockService)
	h.ForgotPasswordHandler(rec, req, httprouter.Params{})

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid request payload")

	assert.NotNil(t, req.Context().Value(middleware.ErrorLogKey))
}

func Test_ForgotPasswordHandler_Fail_ServiceError(t *testing.T) {
	mockService := new(service.MockPasswordResetTokensService)

	jsonPayload := `{"email": "rabs@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewBufferString(jsonPayload))
	rec := httptest.NewRecorder()

	mockService.On("RequestResetPassword", mock.Anything, "rabs@example.com").Return(errors.New("database down"))

	h := NewPasswordResetTokensHandler(mockService)
	h.ForgotPasswordHandler(rec, req, httprouter.Params{})

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "unable to process request")
	assert.NotNil(t, req.Context().Value(middleware.ErrorLogKey))
	mockService.AssertExpectations(t)
}

func Test_ResetPasswordHandler_Pass(t *testing.T) {
	mockService := new(service.MockPasswordResetTokensService)

	jsonPayload := `{"password1": "NewPass123", "password2": "NewPass123"}`
	req := httptest.NewRequest(http.MethodPost, "/reset-password?token=valid-token-rabs", bytes.NewBufferString(jsonPayload))
	rec := httptest.NewRecorder()

	mockService.On("ExecuteResetPassword", mock.Anything, "valid-token-rabs", "NewPass123", "NewPass123").Return(nil)

	h := NewPasswordResetTokensHandler(mockService)
	h.ResetPasswordHandler(rec, req, httprouter.Params{})

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "success")
	mockService.AssertExpectations(t)
}

func Test_ResetPasswordHandler_Fail_TokenMissing(t *testing.T) {
	mockService := new(service.MockPasswordResetTokensService)

	jsonPayload := `{"password1": "NewPass123", "password2": "NewPass123"}`
	// Tanpa query parameter token rabs
	req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewBufferString(jsonPayload))
	rec := httptest.NewRecorder()

	h := NewPasswordResetTokensHandler(mockService)
	h.ResetPasswordHandler(rec, req, httprouter.Params{})

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "token is required")

	assert.Nil(t, req.Context().Value(middleware.ErrorLogKey))
}

func Test_ResetPasswordHandler_Fail_ServiceValidationError(t *testing.T) {
	mockService := new(service.MockPasswordResetTokensService)

	jsonPayload := `{"password1": "NewPass123", "password2": "BedaPass123"}`
	req := httptest.NewRequest(http.MethodPost, "/reset-password?token=valid-token-rabs", bytes.NewBufferString(jsonPayload))
	rec := httptest.NewRecorder()

	mockService.On("ExecuteResetPassword", mock.Anything, "valid-token-rabs", "NewPass123", "BedaPass123").Return(errors.New("password not equal"))

	h := NewPasswordResetTokensHandler(mockService)
	h.ResetPasswordHandler(rec, req, httprouter.Params{})

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "unable to process request")
	assert.NotNil(t, req.Context().Value(middleware.ErrorLogKey))
	mockService.AssertExpectations(t)
}
