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

type UserResponse struct {
	UserId int64  `json:"user_id"`
	Email  string `json:"email"`
}

type UserDetailsJWT struct {
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

type UserHandler struct {
	UserService     domain.UserService
	UserOtpsService domain.UserOtpsService
	CookieConfig    CookieConfig
	JwtSecret       []byte
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword1    string `json:"new_password1"`
	NewPassword2    string `json:"new_password2"`
}

type ResetPasswordRequest struct {
	Email      string `json:"email"`
	ResetToken string `json:"reset_token"`
	Password1  string `json:"password_1"`
	Password2  string `json:"password_2"`
}

func NewUserHandler(userService domain.UserService, userOtps domain.UserOtpsService, cookieConfig *CookieConfig, jwtSecret []byte) *UserHandler {
	return &UserHandler{
		UserService:     userService,
		UserOtpsService: userOtps,
		CookieConfig:    *cookieConfig,
		JwtSecret:       jwtSecret,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var req UserCredentials

	request := json.NewDecoder(r.Body)
	request.DisallowUnknownFields()
	err := request.Decode(&req)
	if err != nil {
		helper.BadResponse(w, http.StatusBadRequest, "invalid request payload")

		middleware.LogWriter(w, "", err)
		return
	}

	if req.Email == "" || req.Password == "" {
		helper.BadResponse(w, http.StatusBadRequest, "email and password are required")

		middleware.LogWriter(w, "", err)
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

		middleware.LogWriter(w, "", err)
		return
	}

	res := UserResponse{
		UserId: user.Id,
		Email:  user.Email,
	}

	helper.GoodResponse(w, http.StatusCreated, "registration successful", res)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var req UserCredentials

	request := json.NewDecoder(r.Body)
	request.DisallowUnknownFields()
	err := request.Decode(&req)
	if err != nil {
		helper.BadResponse(w, http.StatusBadRequest, "invalid request payload")
		middleware.LogWriter(w, "", err)
		return
	}

	if req.Email == "" || req.Password == "" {
		helper.BadResponse(w, http.StatusBadRequest, "email and password are required")

		middleware.LogWriter(w, "", domain.ErrMissingEmailOrPassword)
		return
	}

	ctx := r.Context()

	token, err := h.UserService.Login(ctx, req.Email, req.Password)
	if err != nil {
		helper.BadResponse(w, http.StatusUnauthorized, "invalid email or password")

		middleware.LogWriter(w, "", err)
		return
	}

	userDetails := &UserDetailsJWT{}
	claims, err := helper.DecodeJWTToken(token, userDetails, h.JwtSecret)
	if err != nil || !claims.Valid {
		helper.BadResponse(w, http.StatusUnauthorized, "invalid token or expired")
		if err != nil {
			middleware.LogWriter(w, "invalid token or expired", err)
		} else {
			middleware.LogWriter(w, "", domain.ErrInvalidTokenOrExpired)
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

	helper.GoodResponse(w, http.StatusOK, "login successfully", &UserResponse{
		Email:  userDetails.Email,
		UserId: userDetails.UserId,
	})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	helper.GoodResponse(w, http.StatusOK, "successfully logged out", nil)
}

func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ctx := r.Context()

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
		helper.BadResponse(w, http.StatusBadRequest, "passwords do not match")
		return
	}

	err = h.UserService.ResetPassword(ctx, userRequest.Password1, userRequest.ResetToken)
	if err != nil {
		helper.BadResponse(w, http.StatusInternalServerError, "failed reset password")

		return
	}

	helper.GoodResponse(w, http.StatusOK, "success", nil)
}

func (h *UserHandler) VerifyUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	user, err := middleware.GetUserDetailFromContext(ctx)
	if err != nil {
		middleware.LogWriter(w, "change password auth", err)
		helper.BadResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var userRequest ChangePasswordRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&userRequest); err != nil {
		middleware.LogWriter(w, "change password decode", err)
		helper.BadResponse(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if userRequest.CurrentPassword == "" || userRequest.NewPassword1 == "" || userRequest.NewPassword2 == "" {
		middleware.LogWriter(w, "change password validation", errors.New("all fields are required"))
		helper.BadResponse(w, http.StatusBadRequest, "all fields are required")
		return
	}

	if userRequest.NewPassword1 != userRequest.NewPassword2 {
		middleware.LogWriter(w, "change password validation", errors.New("passwords do not match"))
		helper.BadResponse(w, http.StatusBadRequest, "passwords do not match")
		return
	}

	err = h.UserService.ChangePassword(ctx, user.UserID, userRequest.CurrentPassword, userRequest.NewPassword1)
	if err != nil {
		middleware.LogWriter(w, "change password service", err)
		if errors.Is(err, domain.ErrInvalidCredentials) {
			helper.BadResponse(w, http.StatusBadRequest, "invalid credentials")
		} else {
			helper.BadResponse(w, http.StatusInternalServerError, "change password unsuccessfully")
		}
		return
	}

	helper.GoodResponse(w, http.StatusOK, "change password successfully", nil)
}

func (h *UserHandler) HelloWorld(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	helper.GoodResponse(w, http.StatusOK, "success", map[string]any{
		"head": "berhasil",
		"code": 200,
	})
}
