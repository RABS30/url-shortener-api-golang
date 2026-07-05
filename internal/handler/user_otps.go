package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"shorter-url/internal/domain"
	"shorter-url/internal/helper"

	"github.com/julienschmidt/httprouter"
)

type userOtpsHandler struct {
	UserOtpsService domain.UserOtpsService
}

func NewUserOtpsHandler(userOtpsService domain.UserOtpsService) *userOtpsHandler {
	return &userOtpsHandler{
		UserOtpsService: userOtpsService,
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
		return
	}
	if userRequest.Email == "" || userRequest.OtpCode == "" || userRequest.OtpType == "" {
		helper.BadResponse(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	token, err := h.UserOtpsService.VerifyOTP(ctx, userRequest.OtpCode, userRequest.Email, userRequest.OtpType)
	if err != nil {
		helper.BadResponse(w, http.StatusBadRequest, "failed to verify otp code")
		return
	}

	if userRequest.OtpType == "verification_account" {
		helper.GoodResponse(w, http.StatusOK, "success", nil)
		return
	}
	log.Println(token)
	helper.GoodResponse(w, http.StatusOK, "success", &ResponseBody{Token: token})
}
