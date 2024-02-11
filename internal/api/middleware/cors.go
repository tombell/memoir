package middleware

import (
	"net/http"
)

// CORS is a middleware function that adds the required response headers for
// Cross-Origin Resource Sharing.
func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Header().Add("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))
			w.Header().Add("Access-Control-Expose-Headers", "*")
			w.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

			next.ServeHTTP(w, r)
		})
	}
}
