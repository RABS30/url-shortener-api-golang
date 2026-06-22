package handler

import (
	"encoding/json"
	"net/http"
	"shorter-url/internal/domain"
	"shorter-url/internal/helper"

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

type userHandler struct {
	UserService domain.AuthService
}

func NewUserHandler(UserService domain.AuthService) *userHandler {
	return &userHandler{
		UserService: UserService,
	}
}

func (h *userHandler) Register(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var req userRequest

	request := json.NewDecoder(r.Body)
	request.DisallowUnknownFields()
	err := request.Decode(&req)
	if err != nil {
		helper.BadResponse(w, http.StatusBadRequest, "invalid json format")
		return
	}

	if req.Email == "" || req.Password == "" {
		helper.BadResponse(w, http.StatusBadRequest, "email and password are required")
		return
	}

	ctx := r.Context()
	user, err := h.UserService.Register(ctx, req.Email, req.Password)
	if err != nil {
		helper.BadResponse(w, http.StatusBadRequest, err.Error())

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

		return
	}

	if req.Email == "" || req.Password == "" {
		helper.BadResponse(w, http.StatusBadRequest, "email and password are required")

		return
	}

	ctx := r.Context()

	token, err := h.UserService.Login(ctx, req.Email, req.Password)
	if err != nil {
		helper.BadResponse(w, http.StatusUnauthorized, err.Error())

		return
	}

	cookie := &http.Cookie{
		Name:     "token",
		Value:    token,
		Domain:   "localhost",
		Path:     "/",
		MaxAge:   3600 * 24,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)

	helper.GoodResponse(w, http.StatusOK, "login successfully", nil)
}
