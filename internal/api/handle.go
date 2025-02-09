package api

import (
	"context"
	"net/http"
	"time"

	"github.com/tombell/middle/ware"

	"github.com/tombell/memoir/internal/api/payload"
	"github.com/tombell/memoir/internal/controllers"
)

func rw[In, Out any](fn controllers.ActionFunc[In, Out]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		logger := ware.LoggerFromContext(ctx)

		input, err := payload.Read[In](r)
		if err != nil {
			payload.WriteError(logger, w, err)
			return
		}

		output, err := fn(ctx, input)
		if err != nil {
			payload.WriteError(logger, w, err)
			return
		}

		if err := payload.Write(w, output); err != nil {
			payload.WriteError(logger, w, err)
		}
	})
}

func w[Out any](fn controllers.WriteOnlyActionFunc[Out]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		logger := ware.LoggerFromContext(ctx)

		output, err := fn(ctx)
		if err != nil {
			payload.WriteError(logger, w, err)
			return
		}

		if err := payload.Write(w, output); err != nil {
			payload.WriteError(logger, w, err)
		}
	})
}
