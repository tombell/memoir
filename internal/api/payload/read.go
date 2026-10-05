package payload

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"reflect"

	"github.com/tombell/memoir/internal/errors"
)

// MaxJSONBodyBytes limits JSON writes to 1 MiB. Multipart uploads have their
// own limits and are not subject to this JSON limit.
const MaxJSONBodyBytes int64 = 1 << 20

// Read reads the HTTP request into the output type T. The data read depends on
// the field tags on the type. Requests with JSON fields require application/json.
func Read[T any](r *http.Request) (T, error) {
	op := errors.Op("payload[read]")

	var in T

	if hasJSONFields(reflect.TypeOf(in)) {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			return in, errors.E(op, http.StatusUnsupportedMediaType,
				errors.M{"message": {"Content-Type must be application/json"}})
		}

		tooLarge := func() error {
			return errors.E(op, http.StatusRequestEntityTooLarge,
				errors.M{"message": {"JSON request body must not exceed 1048576 bytes"}})
		}
		if r.ContentLength > MaxJSONBodyBytes {
			return in, tooLarge()
		}

		body, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, MaxJSONBodyBytes))
		if err != nil {
			var sizeError *http.MaxBytesError
			if errors.As(err, &sizeError) {
				return in, tooLarge()
			}
			return in, errors.E(op, http.StatusBadRequest,
				errors.Strf("could not read json: %w", err),
				errors.M{"message": {"Could not read JSON request body"}})
		}

		// Unmarshal requires one complete JSON value and rejects trailing data.
		if err := json.Unmarshal(body, &in); err != nil {
			return in, errors.E(op, http.StatusBadRequest,
				errors.Strf("could not decode json: %w", err),
				errors.M{"message": {"Request body must contain a single valid JSON value"}})
		}
	}

	if err := decode(r, &in); err != nil {
		return in, err
	}

	return in, nil
}

// hasJSONFields distinguishes JSON inputs from query, path, and multipart inputs.
func hasJSONFields(st reflect.Type) bool {
	if st == nil || st.Kind() != reflect.Struct {
		return false
	}

	for i := range st.NumField() {
		if tag, ok := st.Field(i).Tag.Lookup("json"); ok && tag != "-" {
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

	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return errors.E(errors.Op("payload[read]"), http.StatusBadRequest,
			errors.M{"message": {"Invalid query string"}})
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
			if fieldValue.Kind() == reflect.Pointer && fieldValue.Type().Elem().Kind() == reflect.String {
				if query.Has(key) {
					value := query.Get(key)
					fieldValue.Set(reflect.ValueOf(&value))
				}
			} else {
				fieldValue.SetString(query.Get(key))
			}
			continue
		}

		if key, ok := field.Tag.Lookup("file"); ok {
			if file, header, err := r.FormFile(key); err == nil {
				val := &File{File: file, Header: header}
				fieldValue.Set(reflect.ValueOf(val))
			}

			continue
		}
	}

	return nil
}
