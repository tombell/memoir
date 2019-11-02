package api

import (
	"net/http"
)

type contextKey string

func use(h http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middlewares {
		h = m(h)
	}

	return h
}
