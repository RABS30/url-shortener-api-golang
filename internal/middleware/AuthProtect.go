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
				helper.BadResponse(w, http.StatusUnauthorized, "")

				return
			}

			tokenString := cookie.Value

			claims := &Claims{}

			token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
				return []byte(secretKey), nil
			}, jwt.WithValidMethods([]string{"HS256"}))

			if err != nil || !token.Valid {
				http.Error(w, `{"message" : "Unathorized: Token invalid or expired"}`, http.StatusUnauthorized)

				return
			}

			ctx := context.WithValue(r.Context(), UserClaims, claims)
			r = r.WithContext(ctx)

			h(w, r, p)
		}
	}
}

func GetUserIDFromContext(r *http.Request, key any) (int64, error) {
	userContext, ok := r.Context().Value(key).(*Claims)
	if !ok || userContext == nil {
		return 0, fmt.Errorf("claims not found in context")
	}

	userId := userContext.UserID

	if userId == 0 {
		return 0, fmt.Errorf("id is zero or empty")
	}

	return userId, nil
}
