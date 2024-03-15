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
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return in, errors.E(op, errors.Strf("could not read json: %w", err))
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

		if key, ok := field.Tag.Lookup("header"); ok {
			reflect.ValueOf(in).Elem().Field(i).SetString(r.Header.Get(key))
			continue
		}

		if key, ok := field.Tag.Lookup("path"); ok {
			reflect.ValueOf(in).Elem().Field(i).SetString(r.PathValue(key))
			continue
		}

		if key, ok := field.Tag.Lookup("query"); ok {
			reflect.ValueOf(in).Elem().Field(i).SetString(r.URL.Query().Get(key))
			continue
		}

		// TODO: add form file
		// payload.File
	}
}
