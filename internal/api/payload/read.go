package payload

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/tombell/memoir/internal/errors"
)

// Read reads the HTTP request into the output type T. The data read depends on
// specific structs on the type.
// Call Cleanup with the returned input even when decoding fails.
func Read[T any](r *http.Request) (T, error) {
	op := errors.Op("payload[read]")

	var in T

	if r.Header.Get("Content-Type") == "application/json" && !hasFileFields(reflect.TypeOf(in)) {
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return in, errors.E(op, errors.Strf("could not decode json: %w", err))
		}
	}

	if err := decode(r, &in); err != nil {
		return in, err
	}

	return in, nil
}

func hasFileFields(st reflect.Type) bool {
	if st == nil || st.Kind() != reflect.Struct {
		return false
	}
	for i := range st.NumField() {
		if _, ok := st.Field(i).Tag.Lookup("file"); ok {
			return true
		}
	}
	return false
}

// decode reads specific data from the HTTP request based on struct tags found
// on the type T.
func decode[T any](r *http.Request, in T) error {
	st := reflect.TypeOf(in).Elem()
	if st.Kind() != reflect.Struct {
		return nil
	}

	for i := range st.NumField() {
		field := st.Field(i)
		fieldValue := reflect.ValueOf(in).Elem().Field(i)

		if key, ok := field.Tag.Lookup("header"); ok {
			fieldValue.SetString(r.Header.Get(key))
			continue
		}

		if key, ok := field.Tag.Lookup("path"); ok {
			fieldValue.SetString(r.PathValue(key))
			continue
		}

		if key, ok := field.Tag.Lookup("query"); ok {
			fieldValue.SetString(r.URL.Query().Get(key))
			continue
		}

		if key, ok := field.Tag.Lookup("file"); ok {
			// Spill larger files to disk rather than keeping the full upload in memory.
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				return fileError(key, err)
			}

			file, header, err := r.FormFile(key)
			if err != nil {
				return fileError(key, err)
			}
			val := &File{File: file, Header: header}
			fieldValue.Set(reflect.ValueOf(val))

			continue
		}
	}

	return nil
}

func fileError(key string, err error) error {
	status := http.StatusBadRequest
	message := "must be supplied as a valid multipart file"
	if errors.Is(err, http.ErrMissingFile) {
		message = "is required"
	}
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		status = http.StatusRequestEntityTooLarge
		message = "multipart request exceeds the upload limit"
	}

	return errors.E(errors.Op("payload[read]"), err, errors.M{key: {message}}, status)
}
