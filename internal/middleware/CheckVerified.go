package middleware

import (
	"net/http"
	"shorter-url/internal/helper"

	"github.com/julienschmidt/httprouter"
)

func CheckVerifiedUser(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		claims, ok := r.Context().Value(UserClaims).(*Claims)
		if !ok || !claims.IsVerified {
			helper.BadResponse(w, http.StatusForbidden, "verified your account")
			
			return
		}

		next(w, r, p)
	}
}
