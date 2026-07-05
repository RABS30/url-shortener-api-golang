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
	ctx := r.Context()

	userId, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		helper.BadResponse(w, http.StatusUnauthorized, "unauthorized")

		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(err)
		}
		return
	}

	idString := p.ByName("shortUrlId")
	shortUrlId, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		helper.BadResponse(w, http.StatusBadRequest, "invalid short url id")

		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(err)
		}
		return
	}

	listEvent, err := h.Service.FindByShortUrlId(ctx, shortUrlId, userId)
	if err != nil {
		helper.BadResponse(w, http.StatusNotFound, "short url data not found")

		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(err)
		}
		return
	}

	helper.GoodResponse(w, http.StatusOK, "ok", listEvent)
}
