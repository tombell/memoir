package middleware

import (
	"net/http"
)

type ContextKey string

type Middleware func(http.HandlerFunc) http.HandlerFunc

func Use(h http.HandlerFunc, middleware ...Middleware) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}

	return h
}
