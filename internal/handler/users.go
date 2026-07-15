package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"shorter-url/internal/domain"
	"shorter-url/internal/helper"
	"shorter-url/internal/middleware"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter"
)

type UserCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	UserId int64  `json:"user_id"`
	Email  string `json:"email"`
}

type userDetailsJWT struct {
	UserId int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
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
	JwtSecret       []byte
}

func NewUserHandler(userService domain.UserService, userOtps domain.UserOtpsService, cookieConfig *CookieConfig, jwtSecret []byte) *userHandler {
	return &userHandler{
		UserService:     userService,
		UserOtpsService: userOtps,
		CookieConfig:    *cookieConfig,
		JwtSecret:       jwtSecret,
	}
}

func (h *userHandler) Register(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var req UserCredentials

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

	token, err := helper.GenerateJWTToken(jwt.MapClaims{
		"email":    req.Email,
		"otp_type": "verification_account",
	}, h.JwtSecret)

	expiredToken, _ := time.ParseDuration("1m")
	http.SetCookie(w, &http.Cookie{
		Name:     "verify-otp",
		Value:    token,
		Domain:   h.CookieConfig.Domain,
		Path:     h.CookieConfig.Path,
		MaxAge:   int(expiredToken.Seconds()),
		HttpOnly: true,
		Secure:   h.CookieConfig.Secure,
		SameSite: http.SameSiteLaxMode,
	})

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
		UserId: user.Id,
		Email:  user.Email,
	}

	helper.GoodResponse(w, http.StatusCreated, "registration successful", res)
}

func (h *userHandler) Login(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var req UserCredentials

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
			wrapper.WriteError(errors.New("missing email or password"))
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

	userDetails := &userDetailsJWT{}
	claims, err := helper.DecodeJWTToken(token, userDetails, h.JwtSecret)
	if err != nil || !claims.Valid {
		helper.BadResponse(w, http.StatusUnauthorized, "invalid token or expired")
		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			if err != nil {
				wrapper.WriteError(fmt.Errorf("invalid token or expired: %w", err))
			} else {
				wrapper.WriteError(errors.New("invalid token claims or expired"))
			}
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

	helper.GoodResponse(w, http.StatusOK, "login successfully", &userResponse{
		Email:  userDetails.Email,
		UserId: userDetails.UserId,
	})
}

func (h *userHandler) ResetPassword(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ctx := r.Context()

	type ResetPasswordRequest struct {
		Email      string `json:"email"`
		ResetToken string `json:"reset_token"`
		Password1  string `json:"password_1"`
		Password2  string `json:"password_2"`
	}
	var userRequest = &ResetPasswordRequest{}

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

func (h *userHandler) HelloWorld(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	helper.GoodResponse(w, http.StatusOK, "success", map[string]any{
		"head": "berhasil",
		"code": 200,
	})
}

func (h *userHandler) VerifyUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, err := middleware.GetUserDetailFromContext(r.Context())
	if err != nil {
		helper.BadResponse(w, http.StatusUnauthorized, "unauthorized")

		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(fmt.Errorf("verify user: %w", err))
		}
		return
	}

	helper.GoodResponse(w, http.StatusOK, "success", user)
}

func (h *userHandler) ChangePassword(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

}
