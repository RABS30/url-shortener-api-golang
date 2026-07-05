package middleware

import (
	"context"
	"fmt"
	"net/http"
	"shorter-url/internal/domain"
	"shorter-url/internal/helper"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter"
)

type ContextKey string

const UserClaimsKey ContextKey = "userClaimsKey"

type UserPrimaryClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `'json:"email"`
	jwt.RegisteredClaims
}

type authMiddleware struct {
	userRepo  domain.UserRepository
	secretKey []byte
}

func NewAuthMiddleware(userRepo domain.UserRepository, secretKey []byte) *authMiddleware {
	return &authMiddleware{
		userRepo:  userRepo,
		secretKey: secretKey,
	}
}

func (m *authMiddleware) Authenticate(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		cookie, err := r.Cookie("token")
		if err != nil {
			helper.BadResponse(w, http.StatusUnauthorized, "unathorized")

			if wrapper, ok := w.(*LogResponseWriter); ok {
				wrapper.WriteError((fmt.Errorf("token not found: %v", err)))
			}
			return
		}

		tokenString := cookie.Value
		userDetailClaims := &UserPrimaryClaims{}

		token, err := jwt.ParseWithClaims(tokenString, userDetailClaims, func(t *jwt.Token) (any, error) {
			return []byte(m.secretKey), nil
		})
		if err != nil || !token.Valid {
			helper.BadResponse(w, http.StatusUnauthorized, "invalid token or expired")

			if wrapper, ok := w.(*LogResponseWriter); ok {
				wrapper.WriteError(fmt.Errorf("invalid token or expired: %w", err))
			}

			return
		}

		ctx := context.WithValue(r.Context(), UserClaimsKey, userDetailClaims)
		r = r.WithContext(ctx)

		h(w, r, p)
	}
}

func (m *authMiddleware) VerifiedOnly(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		claims, ok := r.Context().Value(UserClaimsKey).(*UserPrimaryClaims)
		if !ok {
			helper.BadResponse(w, http.StatusUnauthorized, "unauthorized")

			if wrapper, ok := w.(*LogResponseWriter); ok {
				wrapper.WriteError(fmt.Errorf("token not found: %t", ok))
			}

			return
		}

		user, err := m.userRepo.FindById(r.Context(), claims.UserID)
		if err != nil || !user.IsVerified {
			helper.BadResponse(w, http.StatusForbidden, "account not verified")

			if wrapper, ok := w.(*LogResponseWriter); ok {
				wrapper.WriteError(err)
			}
			return
		}

		h(w, r, p)
	}
}

func (m *authMiddleware) GuestOnly(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		cookie, _ := r.Cookie("token")

		if cookie != nil {
			token := TokenIsValid(string(m.secretKey), cookie.Value)
			if token {
				helper.BadResponse(w, http.StatusForbidden, "already authenticated")
				if wrapper, ok := w.(*LogResponseWriter); ok {
					wrapper.WriteError(fmt.Errorf("user already authenticated: %s", cookie.Value))
				}
				return
			}
		}

		h(w, r, p)
	}
}

func GetUserIDFromContext(ctx context.Context) (int64, error) {
	userContext, ok := ctx.Value(UserClaimsKey).(*UserPrimaryClaims)
	if !ok || userContext == nil {
		return 0, fmt.Errorf("user data not found")
	}

	userId := userContext.UserID

	if userId == 0 {
		return 0, fmt.Errorf("user id not found")
	}

	return userId, nil
}

func TokenIsValid(secretKey string, tokenString string) bool {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return false
	} else {
		return token != nil && token.Valid
	}
}
