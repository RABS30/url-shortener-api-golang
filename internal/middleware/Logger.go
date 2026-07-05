package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type customSlogHandler struct {
	slog.Handler
}

func (h *customSlogHandler) Handle(ctx context.Context, r slog.Record) error {
	userID, _ := GetUserIDFromContext(ctx)
	if userID != 0 {
		r.AddAttrs(slog.Int64("user_id", userID))
	}

	return h.Handler.Handle(ctx, r)
}

func InitLogger() {
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "password" || a.Key == "token" {
				return slog.String(a.Key, "********")
			}
			return a
		},
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, opts)
	customHandler := &customSlogHandler{jsonHandler}

	slog.SetDefault(slog.New(customHandler))
}

type ctxKey string

type LogResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Error      error
}

func (r *LogResponseWriter) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *LogResponseWriter) WriteError(err error) {
	r.Error = err
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		wrapperResponseWriter := &LogResponseWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
			Error:          nil,
		}

		next.ServeHTTP(wrapperResponseWriter, r)

		ctx := r.Context()

		logAttrs := []slog.Attr{
			slog.Group("request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_ip", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			),
			slog.Group("response",
				slog.Int("status", wrapperResponseWriter.StatusCode),
				slog.String("latency", time.Since(startTime).String()),
			),
		}
		if wrapperResponseWriter.Error != nil {
			logAttrs = append(logAttrs,
				slog.Group("error",
					slog.Any("message", wrapperResponseWriter.Error)))
		}
		requestID := GetRequestIDFromContext(ctx)
		if requestID != "" {
			logAttrs = append(logAttrs, slog.String("request_id", requestID))
		}

		msg := "HTTP Request"
		level := slog.LevelInfo

		switch {
		case wrapperResponseWriter.StatusCode >= 500:
			level = slog.LevelError
		case wrapperResponseWriter.StatusCode >= 400:
			level = slog.LevelWarn
		}

		slog.LogAttrs(ctx, level, msg, logAttrs...)
	})
}
