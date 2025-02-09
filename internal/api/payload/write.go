package payload

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/tombell/memoir/internal/errors"
)

// Write writes the data to the HTTP response depending on struct tags on the
// type T.
func Write[T any](w http.ResponseWriter, out T) error {
	op := errors.Op("payload[write]")

	status := http.StatusOK
	if sc, ok := any(out).(StatusCoder); ok {
		status = sc.StatusCode()
	}

	encode(w, out)

	// TODO: move this and the JSON encoding into the encode method?
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(out); err != nil {
		return errors.E(op, errors.Strf("could not encode json: %w", err))
	}

	return nil
}

// WriteError writes the given error to the HTTP response. Depending on the
// interfaces implemented by the error, different status codes and/or error
// messages may be written.
func WriteError(logger *slog.Logger, w http.ResponseWriter, err error) {
	resp := ErrorResponse{
		Errors: errors.M{"message": []string{"something went wrong"}},
		status: http.StatusInternalServerError,
	}

	if cr, ok := err.(ClientReporter); ok {
		resp.status = cr.Status()
		resp.Errors = cr.Message()
	}

	if resp.status >= http.StatusInternalServerError {
		logger.Error("something went wrong", "err", err)
	}

	if err := Write(w, &resp); err != nil {
		logger.Error("could not write error response", "err", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}
}

// encode writes specifc data to the HTTP response based on struct tags found
// on the type T.
func encode[T any](w http.ResponseWriter, out T) {
	st := reflect.TypeOf(out).Elem()
	if st.Kind() != reflect.Struct {
		return
	}

	for i := range st.NumField() {
		field := st.Field(i)

		if key, ok := field.Tag.Lookup("header"); ok {
			val := reflect.ValueOf(out).Elem().Field(i).String()
			w.Header().Add(key, val)
			continue
		}
	}
}
