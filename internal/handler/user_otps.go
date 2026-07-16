package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"shorter-url/internal/domain"
	"shorter-url/internal/helper"
	"shorter-url/internal/middleware"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter"
)

type userOtpsHandler struct {
	UserOtpsService domain.UserOtpsService
	SecretKey       []byte
}

type SessionOtpPageClaims struct {
	Email   string `json:"email"`
	OtpType string `json:"otp_type"`
	jwt.RegisteredClaims
}

func NewUserOtpsHandler(userOtpsService domain.UserOtpsService, secretKey []byte) *userOtpsHandler {
	return &userOtpsHandler{
		UserOtpsService: userOtpsService,
		SecretKey:       secretKey,
	}
}

func (h *userOtpsHandler) RequestOTP(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ctx := r.Context()

	type RequestBody struct {
		Email   string `json:"email"`
		OtpType string `json:"otp_type"`
	}
	userRequest := &RequestBody{}

	request := json.NewDecoder(r.Body)
	request.DisallowUnknownFields()
	err := request.Decode(&userRequest)
	if err != nil {
		helper.BadResponse(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if userRequest.Email == "" || userRequest.OtpType == "" {
		helper.BadResponse(w, http.StatusBadRequest, "email or otp_type are required")
		return
	}

	err = h.UserOtpsService.SendOTP(ctx, userRequest.Email, userRequest.OtpType)
	if err != nil {
		helper.BadResponse(w, http.StatusInternalServerError, "failed send otp code")
		return
	}

	helper.GoodResponse(w, http.StatusOK, "success", nil)
}

func (h *userOtpsHandler) VerifyOTP(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ctx := r.Context()

	type RequestBody struct {
		Email   string `json:"email"`
		OtpType string `json:"otp_type"`
		OtpCode string `json:"otp_code"`
	}
	var userRequest = &RequestBody{}

	type ResponseBody struct {
		Token string `json:"token"`
	}

	request := json.NewDecoder(r.Body)
	request.DisallowUnknownFields()
	err := request.Decode(&userRequest)
	if err != nil {
		helper.BadResponse(w, http.StatusBadRequest, "invalid request payload")

		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(fmt.Errorf("Verify  OTP code: %w", err))
		}
		return
	}
	if userRequest.Email == "" || userRequest.OtpCode == "" || userRequest.OtpType == "" {
		helper.BadResponse(w, http.StatusBadRequest, "invalid request payload")

		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(fmt.Errorf("Verify  OTP code: %w", err))
		}
		return
	}

	token, err := h.UserOtpsService.VerifyOTP(ctx, userRequest.OtpCode, userRequest.Email, userRequest.OtpType)
	if err != nil {
		helper.BadResponse(w, http.StatusBadRequest, "invalid code")

		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(fmt.Errorf("Verify  OTP code: %w", err))
		}
		return
	}

	if userRequest.OtpType == "verification_account" {
		helper.GoodResponse(w, http.StatusOK, "verify successfully", nil)

		return
	}

	helper.GoodResponse(w, http.StatusOK, "verify successfully", &ResponseBody{Token: token})
}

func (h *userOtpsHandler) VerifySessionOtpPage(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	claims := &SessionOtpPageClaims{}

	tokenString, err := r.Cookie("verify-otp")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			helper.BadResponse(w, http.StatusForbidden, "forbidden: missing token")
			return
		}
		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(err)
		}
		helper.BadResponse(w, http.StatusForbidden, "forbidden")
		return
	}

	token, err := jwt.ParseWithClaims(tokenString.Value, claims, func(t *jwt.Token) (any, error) {
		return []byte(h.SecretKey), nil
	})
	if err != nil || !token.Valid || claims.OtpType != "verification_account" {
		helper.BadResponse(w, http.StatusForbidden, "invalid token or expired")

		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(fmt.Errorf("invalid token or expired: %w", err))
		}

		return
	}

	type responseBody struct {
		Email   string `json:"email"`
		OtpType string `json:"otp_type"`
	}

	helper.GoodResponse(w, http.StatusOK, "ok", &responseBody{
		Email:   claims.Email,
		OtpType: claims.OtpType,
	})

}
