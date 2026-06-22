package handler

import (
	"encoding/json"
	"net/http"
	"shorter-url/internal/domain"
	"shorter-url/internal/helper"
	"shorter-url/internal/middleware"
	"time"

	"github.com/julienschmidt/httprouter"
)

type ShortUrlResponse struct {
	Id          int64     `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalUrl string    `json:"original_url"`
	ExpiredAt   time.Time `json:"expired_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type shortUrlHandler struct {
	Service domain.ShortUrlsService
}

func NewShortUrlHandler(service domain.ShortUrlsService) *shortUrlHandler {
	return &shortUrlHandler{
		Service: service,
	}
}

func (s *shortUrlHandler) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userId := middleware.GetUserIDFromCookie(r, middleware.UserIDKey)
	if userId == 0 {
		helper.BadResponse(w, http.StatusUnauthorized, "Unauthorized")

		return
	}

	var inputData struct {
		OriginalUrl string `json:"original_url"`
	}

	inputRequest := json.NewDecoder(r.Body)
	inputRequest.DisallowUnknownFields()
	if inputRequest.Decode(&inputData) != nil {
		helper.BadResponse(w, http.StatusBadRequest, "invalid json format")

		return
	}

	ctx := r.Context()
	expiredAt := time.Now().AddDate(0, 1, 0)

	result, err := s.Service.CreateShortUrl(ctx, userId, inputData.OriginalUrl, expiredAt)
	if err != nil {
		helper.BadResponse(w, http.StatusBadRequest, "failed to create short code")

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	data := &ShortUrlResponse{
		Id:          result.Id,
		ShortCode:   result.ShortCode,
		OriginalUrl: result.OriginalUrl,
		ExpiredAt:   result.ExpiredAt,
		CreatedAt:   result.CreatedAt,
	}

	json.NewEncoder(w).Encode(map[string]any{
		"message": "Short code created successfuly",
		"data":    data,
	})
}

func (s *shortUrlHandler) AccessShortCode(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ctx := r.Context()
	shortCode := p.ByName("shortCode")

	result, err := s.Service.GetShortUrlByShortCode(ctx, shortCode)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"message": "Short code not found",
		})

		return
	}

	http.Redirect(w, r, result.OriginalUrl, http.StatusFound)
}
