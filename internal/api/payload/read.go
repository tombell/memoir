package payload

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/tombell/memoir/internal/errors"
)

func Read[T any](r *http.Request) (T, error) {
	op := errors.Op("payload[read]")

	var in T

	if r.Header.Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return in, errors.E(op, errors.Strf("could not decode json: %w", err))
		}
	}

	decode(r, &in)

	return in, nil
}

func decode[T any](r *http.Request, in T) {
	st := reflect.TypeOf(in).Elem()
	if st.Kind() != reflect.Struct {
		return
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
			if file, header, err := r.FormFile(key); err == nil {
				val := &File{File: file, Header: header}
				fieldValue.Set(reflect.ValueOf(val))
			}

			continue
		}
	}
}
