package payload

import (
	"bytes"
	"encoding/json"
	"io"
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

	if status == http.StatusNoContent {
		encode(w, out)
		w.WriteHeader(status)
		return nil
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(out); err != nil {
		return errors.E(op, errors.Strf("could not encode json: %w", err))
	}

	encode(w, out)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	n, err := w.Write(body.Bytes())
	if err == nil && n != body.Len() {
		err = io.ErrShortWrite
	}
	if err != nil {
		return &responseWriteError{errors.E(op, errors.Strf("could not write json: %w", err))}
	}

	return nil
}

// WriteError writes the given error to the HTTP response. Depending on the
// interfaces implemented by the error, different status codes and/or error
// messages may be written.
func WriteError(logger *slog.Logger, w http.ResponseWriter, err error) {
	if logger == nil {
		logger = slog.Default()
	}

	var writeErr *responseWriteError
	if errors.As(err, &writeErr) {
		logger.Error("could not write response", "err", err)
		return
	}
	if response, ok := w.(interface{ ResponseCommitted() bool }); ok && response.ResponseCommitted() {
		logger.Error("response already committed", "err", err)
		return
	}

	resp := ErrorResponse{
		Errors: errors.M{"message": []string{"something went wrong"}},
		status: http.StatusInternalServerError,
	}

	var e *errors.Error
	if errors.As(err, &e) {
		resp.status = e.Status()
		resp.Errors = e.Message()
	}

	if resp.status >= http.StatusInternalServerError {
		logger.Error("something went wrong", "err", err)
	}

	if err := Write(w, &resp); err != nil {
		logger.Error("could not write error response", "err", err)
	}
}

// responseWriteError indicates that headers have been committed and the writer
// failed. Only encoding errors can still be replaced with an error response.
type responseWriteError struct {
	error
}

func (e *responseWriteError) Unwrap() error {
	return e.error
}

// encode writes specifc data to the HTTP response based on struct tags found
// on the type T.
func encode[T any](w http.ResponseWriter, out T) {
	value := reflect.ValueOf(out)
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return
	}
	st := value.Type()

	for i := range st.NumField() {
		field := st.Field(i)

		if key, ok := field.Tag.Lookup("header"); ok {
			val := value.Field(i).String()
			w.Header().Add(key, val)
			continue
		}
	}
}
