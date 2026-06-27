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

const UserIDKey contextKey = "userID"

type Claims struct {
	UserID int64 `json:"user_id"`
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

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			r = r.WithContext(ctx)

			h(w, r, p)
		}
	}
}

func GetUserIDFromContext(r *http.Request, key any) (int64, error) {
	userIDContext := r.Context().Value(key)
	if userIDContext == nil {
		return 0, fmt.Errorf("id not found")
	}

	if userID, ok := userIDContext.(int64); ok {
		return userID, nil
	}

	return 0, fmt.Errorf("invalid id")
}
