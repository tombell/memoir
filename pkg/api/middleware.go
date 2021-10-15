package api

import (
	"net/http"
)

type contextKey string

type middleware func(http.HandlerFunc) http.HandlerFunc

func use(h http.HandlerFunc, middlewares ...middleware) http.HandlerFunc {
	for _, m := range middlewares {
		h = m(h)
	}

	return h
}
