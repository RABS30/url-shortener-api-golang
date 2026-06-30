package handler

import (
	"encoding/json"
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
	UserService  domain.AuthService
	CookieConfig CookieConfig
}

func NewUserHandler(userService domain.AuthService, cookieConfig *CookieConfig) *userHandler {
	return &userHandler{
		UserService:  userService,
		CookieConfig: *cookieConfig,
	}
}

func (h *userHandler) Register(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var req userRequest

	request := json.NewDecoder(r.Body)
	request.DisallowUnknownFields()
	err := request.Decode(&req)
	if err != nil {
		helper.BadResponse(w, http.StatusBadRequest, "invalid request payload")

		if wrapper, ok := w.(*middleware.ResponseWriterWrapper); ok {
			wrapper.WriteError(err.Error())
		}
		return
	}

	if req.Email == "" || req.Password == "" {
		helper.BadResponse(w, http.StatusBadRequest, "email and password are required")

		if wrapper, ok := w.(*middleware.ResponseWriterWrapper); ok {
			wrapper.WriteError("email and password are required")
		}
		return
	}

	ctx := r.Context()
	user, err := h.UserService.Register(ctx, req.Email, req.Password)
	if err != nil {
		if err.Error() == "email already registered" {
			helper.BadResponse(w, http.StatusConflict, "email already registered")
		} else {
			helper.BadResponse(w, http.StatusInternalServerError, "register failed")
		}

		if wrapper, ok := w.(*middleware.ResponseWriterWrapper); ok {
			wrapper.WriteError(err.Error())
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
		helper.BadResponse(w, http.StatusBadRequest, "invalid json format")

		if wrapper, ok := w.(*middleware.ResponseWriterWrapper); ok {
			wrapper.WriteError(err.Error())
		}
		return
	}

	if req.Email == "" || req.Password == "" {
		helper.BadResponse(w, http.StatusBadRequest, "email and password are required")

		if wrapper, ok := w.(*middleware.ResponseWriterWrapper); ok {
			wrapper.WriteError("email and password are required")
		}
		return
	}

	ctx := r.Context()

	token, err := h.UserService.Login(ctx, req.Email, req.Password)
	if err != nil {
		helper.BadResponse(w, http.StatusUnauthorized, "invalid email or password")

		if wrapper, ok := w.(*middleware.ResponseWriterWrapper); ok {
			wrapper.WriteError(err.Error())
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
