package api

import (
	"net/http"
)

func (s *Server) auth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)
		key := r.Header.Get("API-Write-Key")

		if key == "" {
			s.services.Logger.Printf("[%s] error=missing api key", rid)
			w.WriteHeader(http.StatusUnauthorized)
		}

		if key != s.services.Config.API.WriteKey {
			s.services.Logger.Printf("[%s] error=invalid api key", rid)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		h(w, r)
	}
}
