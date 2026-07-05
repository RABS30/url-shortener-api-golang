package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const requestIDCtxKey ctxKey = "request_id"
const requestIDHeader = "X-Request-ID"

func RequestID(handle http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetOrCreateRequestID(r)

		requestIDCtx := context.WithValue(r.Context(), requestIDCtxKey, requestID)
		r = r.WithContext(requestIDCtx)

		w.Header().Set(requestIDHeader, requestID)

		handle.ServeHTTP(w, r)
	})
}

func GetOrCreateRequestID(r *http.Request) string {
	requestID := r.Header.Get(requestIDHeader)

	if requestID == "" {
		requestID = uuid.NewString()
	}

	return requestID
}

func GetRequestIDFromContext(ctx context.Context) string {
	requestID, ok := ctx.Value(requestIDCtxKey).(string)
	if !ok {
		return ""
	}

	return requestID
}
