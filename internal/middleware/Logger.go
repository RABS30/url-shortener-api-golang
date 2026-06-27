package middleware

import (
	"log"
	"net/http"
	"time"
)

const ErrorLogKey contextKey = "errorDetails"

type responseWriterWrapper struct {
	http.ResponseWriter
	StatusCode int
}

func (r *responseWriterWrapper) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		responseWriter := &responseWriterWrapper{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		next.ServeHTTP(responseWriter, r)

		var errorDetail string = "-"
		if responseWriter.StatusCode >= 400 {
			if err, ok := r.Context().Value(ErrorLogKey).(error); ok && err != nil {
				errorDetail = "error: " + err.Error()
			} else {
				errorDetail = "no detailed error"
			}
		}

		log.Printf(
			"| %d | %s | %s | %s | %s | %s |",
			responseWriter.StatusCode,
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			time.Since(startTime),
			errorDetail,
		)
	})
}
