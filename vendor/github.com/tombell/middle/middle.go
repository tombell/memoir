package middle

import "net/http"

// Use creates a middleware chain composed of the given middleware functions and
// returns a function that takes a http.Handler, usually the main handler for
// the HTTP route.
func Use(middleware ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for i := len(middleware) - 1; i >= 0; i-- {
			next = middleware[i](next)
		}

		return next
	}
}
