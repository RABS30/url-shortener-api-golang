package handler

import (
	"net/http"
	"shorter-url/internal/domain"
	"shorter-url/internal/helper"
	"shorter-url/internal/middleware"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type clickEventHandler struct {
	Service domain.ClickEventService
}

func NewClickEventHandler(service domain.ClickEventService) *clickEventHandler {
	return &clickEventHandler{
		Service: service,
	}
}

func (h *clickEventHandler) FindByShortUrlId(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userID, err := middleware.GetUserIDFromContext(r, middleware.UserIDKey)
	if err != nil {
		helper.BadResponse(w, http.StatusUnauthorized, "Unauthorized")

		return
	}

	ctx := r.Context()

	idString := p.ByName("shortUrlId")
	shortUrlId, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		helper.BadResponse(w, http.StatusBadRequest, "invalid short url id")

		return
	}

	listEvent, err := h.Service.FindByShortUrlId(ctx, shortUrlId, userID)
	if err != nil {
		helper.BadResponse(w, http.StatusNotFound, "click event not found")

		return
	}

	helper.GoodResponse(w, http.StatusOK, "ok", listEvent)
}
