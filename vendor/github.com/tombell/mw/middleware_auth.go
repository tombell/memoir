package mw

import (
	"log"
	"net/http"
)

func Auth(logger *log.Logger, token string) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			rid := FindRequestID(r)
			key := r.Header.Get("API-Token")

			if key == "" {
				logger.Printf("[%s] error=missing api key", rid)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if key != token {
				logger.Printf("[%s] error=invalid api key", rid)
				w.WriteHeader(http.StatusForbidden)
				return
			}

			h(w, r)
		}
	}
}
