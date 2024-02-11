package middleware

import (
	"net/http"
)

// ContextKey is used to define keys for values added to the request context.
type ContextKey string

// Use creates a middleware chain composed of the given middleware functions and
// returns a function that takes a http.Handler, usually the main handler for
// the route.
func Use(middleware ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for _, m := range middleware {
			next = m(next)
		}

		return next
	}
}
