package middleware

import (
	"errors"
	"net/http"
	"shorter-url/internal/helper"

	"github.com/julienschmidt/httprouter"
)

func VerifiedUserOnly(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		claims, ok := r.Context().Value(UserClaims).(*Claims)
		if !ok || !claims.IsVerified {
			helper.BadResponse(w, http.StatusForbidden, "account not verified")

			if wrapper, ok := w.(*LogResponseWriter); ok {
				wrapper.WriteError(errors.New("account not verified"))
			}
			return
		}

		next(w, r, p)
	}
}
