package api

import (
	"net/http"
	"time"
)

func (s *Server) instruments(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		start := time.Now().UTC()
		method := r.Method
		path := r.URL.Path
		addr := r.RemoteAddr

		s.services.Logger.Printf("rid=%s method=%s path=%s ip=%s\n", rid, method, path, addr)
		h(w, r)
		s.services.Logger.Printf("rid=%s method=%s path=%s ip=%s elapsed=%s\n", rid, method, path, addr, time.Since(start))
	}
}
