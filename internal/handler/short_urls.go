package handler

import (
	"context"
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
	Service    domain.ShortUrlsService
	ClickEvent domain.ClickEventService
}

func NewShortUrlHandler(service domain.ShortUrlsService, clickEvent domain.ClickEventService) *shortUrlHandler {
	return &shortUrlHandler{
		Service:    service,
		ClickEvent: clickEvent,
	}
}

func (s *shortUrlHandler) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userId, err := middleware.GetUserIDFromContext(r, middleware.UserIDKey)
	if err != nil {
		errorCtx := context.WithValue(r.Context(), middleware.ErrorLogKey, err)
		*r = *r.WithContext(errorCtx)

		helper.BadResponse(w, http.StatusInternalServerError, "server broken")
		return
	}
	if userId == 0 {
		helper.BadResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var inputData struct {
		OriginalUrl string `json:"original_url"`
	}

	inputRequest := json.NewDecoder(r.Body)
	inputRequest.DisallowUnknownFields()
	if err := inputRequest.Decode(&inputData); err != nil {
		errorCtx := context.WithValue(r.Context(), middleware.ErrorLogKey, err)
		*r = *r.WithContext(errorCtx)

		helper.BadResponse(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	ctx := r.Context()
	expiredAt := time.Now().AddDate(0, 1, 0)

	result, err := s.Service.CreateShortUrl(ctx, userId, inputData.OriginalUrl, expiredAt)
	if err != nil {
		errorCtx := context.WithValue(r.Context(), middleware.ErrorLogKey, err)
		*r = *r.WithContext(errorCtx)

		helper.BadResponse(w, http.StatusBadRequest, "failed to create short code")
		return
	}

	data := &ShortUrlResponse{
		Id:          result.Id,
		ShortCode:   result.ShortCode,
		OriginalUrl: result.OriginalUrl,
		ExpiredAt:   result.ExpiredAt,
		CreatedAt:   result.CreatedAt,
	}

	helper.GoodResponse(w, http.StatusCreated, "Short code created successfully", data)
}

func (s *shortUrlHandler) AccessShortCode(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ctx := r.Context()
	shortCode := p.ByName("shortCode")

	result, err := s.Service.GetShortUrlByShortCode(ctx, shortCode)
	if err != nil {
		errorCtx := context.WithValue(r.Context(), middleware.ErrorLogKey, err)
		*r = *r.WithContext(errorCtx)

		helper.BadResponse(w, http.StatusBadRequest, "short code not found")
		return
	}

	ipAddress := r.RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ipAddress = xff
	}

	userAgent := r.UserAgent()

	referer := r.Referer()
	if referer == "" {
		referer = "Direct"
	}

	metadataUser := &domain.ClickEvent{
		ShortUrlId: result.Id,
		IpAddress:  ipAddress,
		UserAgent:  userAgent,
		Referer:    referer,
	}

	_, err = s.ClickEvent.Create(ctx, metadataUser)
	if err != nil {
		errorCtx := context.WithValue(r.Context(), middleware.ErrorLogKey, err)
		*r = *r.WithContext(errorCtx)

		helper.BadResponse(w, http.StatusInternalServerError, "failed to record metric data")
		return
	}

	http.Redirect(w, r, result.OriginalUrl, http.StatusFound)
}
