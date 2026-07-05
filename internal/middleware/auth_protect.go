package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"shorter-url/internal/helper"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter"
)

type contextKey string

const UserClaims contextKey = "userClaims"

type Claims struct {
	UserID     int64  `json:"user_id"`
	Email      string `json:"email"`
	IsVerified bool   `json:"is_verified"`
	jwt.RegisteredClaims
}

func AuthMiddleware(secretKey string) func(httprouter.Handle) httprouter.Handle {
	if secretKey == "" {
		log.Fatal("secret key not found")
	}

	return func(h httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

			cookie, err := r.Cookie("token")
			if err != nil {
				helper.BadResponse(w, http.StatusUnauthorized, "unauthorized")

				if wrapper, ok := w.(*LogResponseWriter); ok {
					wrapper.WriteError(fmt.Errorf("auth middleware: %s", err.Error()))
				}
				return
			}

			tokenString := cookie.Value

			claims := &Claims{}

			token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
				return []byte(secretKey), nil
			}, jwt.WithValidMethods([]string{"HS256"}))

			if err != nil || !token.Valid {
				helper.BadResponse(w, http.StatusUnauthorized, "invalid token or expired")

				if wrapper, ok := w.(*LogResponseWriter); ok {
					wrapper.WriteError(err)
				}
				return
			}

			ctx := context.WithValue(r.Context(), UserClaims, claims)
			r = r.WithContext(ctx)

			h(w, r, p)
		}
	}
}
