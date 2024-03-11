package ware

import (
	"net/http"
	"strings"
)

type CORSOptions struct {
	AllowedOrigins   []string
	AllowedHeaders   []string
	ExposeHeaders    []string
	AllowedMethods   []string
	AllowCredentials bool
}

// CORS is a middleware function that adds the required response headers for
// Cross-Origin Resource Sharing.
func CORS(cfg CORSOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowedOrigin := ""

			if len(cfg.AllowedOrigins) > 0 {
				origin := r.Header.Get("Origin")

				for _, o := range cfg.AllowedOrigins {
					if o == origin {
						allowedOrigin = o
						break
					}
				}
			} else {
				allowedOrigin = "*"
			}

			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)

			if len(cfg.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
			}

			if len(cfg.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
			}

			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
