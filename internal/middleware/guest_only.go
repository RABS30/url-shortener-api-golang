package middleware

import (
	"errors"
	"log"
	"net/http"
	"shorter-url/internal/helper"

	"github.com/julienschmidt/httprouter"
)

func GuestOnly(secretKey string) func(httprouter.Handle) httprouter.Handle {
	if secretKey == "" {
		log.Fatal("secret key not found")
	}
	return func(next httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			token, _ := r.Cookie("token")
			if token != nil {
				valid := TokenIsValid(secretKey, token.Value)
				if valid {
					helper.BadResponse(w, http.StatusBadRequest, "already authenticated")

					if wrapper, ok := w.(*LogResponseWriter); ok {
						wrapper.WriteError(errors.New("user already authenticated"))
					}

					return
				}
			}

			next(w, r, p)
		}
	}
}
