package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"shorter-url/internal/domain"
	"shorter-url/internal/helper"
	"shorter-url/internal/middleware"

	"github.com/julienschmidt/httprouter"
)

type userRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	Id    int64  `json:"id"`
	Email string `json:"email"`
}

type CookieConfig struct {
	Domain string
	MaxAge int
	Secure bool
	Path   string
}

type userHandler struct {
	UserService     domain.UserService
	UserOtpsService domain.UserOtpsService
	CookieConfig    CookieConfig
}

func NewUserHandler(userService domain.UserService, userOtps domain.UserOtpsService, cookieConfig *CookieConfig) *userHandler {
	return &userHandler{
		UserService:     userService,
		UserOtpsService: userOtps,
		CookieConfig:    *cookieConfig,
	}
}

func (h *userHandler) Register(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var req userRequest

	request := json.NewDecoder(r.Body)
	request.DisallowUnknownFields()
	err := request.Decode(&req)
	if err != nil {
		helper.BadResponse(w, http.StatusBadRequest, "invalid request payload")

		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(err)
		}
		return
	}

	if req.Email == "" || req.Password == "" {
		helper.BadResponse(w, http.StatusBadRequest, "email and password are required")

		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(err)
		}
		return
	}

	ctx := r.Context()
	user, err := h.UserService.Register(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyRegistered) {
			helper.BadResponse(w, http.StatusConflict, "email already registered")
		} else {
			helper.BadResponse(w, http.StatusInternalServerError, "register failed")
		}

		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(err)
		}
		return
	}

	res := userResponse{
		Id:    user.Id,
		Email: user.Email,
	}

	helper.GoodResponse(w, http.StatusCreated, "registration successful", res)
}

func (h *userHandler) Login(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var req userRequest

	request := json.NewDecoder(r.Body)
	request.DisallowUnknownFields()
	err := request.Decode(&req)
	if err != nil {
		helper.BadResponse(w, http.StatusBadRequest, "invalid request payload")

		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(err)
		}
		return
	}

	if req.Email == "" || req.Password == "" {
		helper.BadResponse(w, http.StatusBadRequest, "email and password are required")

		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(err)
		}
		return
	}

	ctx := r.Context()

	token, err := h.UserService.Login(ctx, req.Email, req.Password)
	if err != nil {
		helper.BadResponse(w, http.StatusUnauthorized, "invalid email or password")

		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(err)
		}
		return
	}

	cookie := &http.Cookie{
		Name:     "token",
		Value:    token,
		Domain:   h.CookieConfig.Domain,
		Path:     h.CookieConfig.Path,
		MaxAge:   h.CookieConfig.MaxAge,
		HttpOnly: true,
		Secure:   h.CookieConfig.Secure,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)

	helper.GoodResponse(w, http.StatusOK, "login successfully", nil)
}

func (h *userHandler) ResetPassword(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ctx := r.Context()

	type RequestBody struct {
		Email      string `json:"email"`
		ResetToken string `json:"reset_token"`
		Password1  string `json:"password_1"`
		Password2  string `json:"password_2"`
	}
	var userRequest = &RequestBody{}

	request := json.NewDecoder(r.Body)
	request.DisallowUnknownFields()
	err := request.Decode(&userRequest)
	if err != nil {
		helper.BadResponse(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if userRequest.Email == "" || userRequest.ResetToken == "" || userRequest.Password1 == "" || userRequest.Password2 == "" {
		helper.BadResponse(w, http.StatusBadRequest, "all fields are required")
		return
	}

	if userRequest.Password1 != userRequest.Password2 {
		helper.BadResponse(w, http.StatusBadRequest, "password1 and password2 do not macth")
		return
	}

	err = h.UserService.ResetPassword(ctx, userRequest.Password1, userRequest.ResetToken)
	if err != nil {
		helper.BadResponse(w, http.StatusInternalServerError, "failed reset password")

		return
	}

	helper.GoodResponse(w, http.StatusOK, "success", nil)
}

func (h *userHandler) ChangePassword(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

}
