package payload

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"reflect"

	"github.com/tombell/memoir/internal/errors"
)

// Read reads the HTTP request into the output type T. The data read depends on
// specific structs on the type.
func Read[T any](r *http.Request) (T, error) {
	op := errors.Op("payload[read]")

	var in T

	media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if (r.Method == http.MethodPost || r.Method == http.MethodPatch) && media == "application/json" {
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&in); err != nil {
			return in, errors.E(op, http.StatusBadRequest)
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			return in, errors.E(op, http.StatusBadRequest)
		}
	} else if r.Method == http.MethodPost || r.Method == http.MethodPatch {
		if media != "multipart/form-data" {
			return in, errors.E(op, http.StatusUnsupportedMediaType)
		}
	}

	if err := decode(r, &in); err != nil {
		return in, err
	}

	return in, nil
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
			if file, header, err := r.FormFile(key); err == nil {
				val := &File{File: file, Header: header}
				fieldValue.Set(reflect.ValueOf(val))
			} else if !errors.Is(err, http.ErrMissingFile) {
				var tooLarge *http.MaxBytesError
				if errors.As(err, &tooLarge) {
					return errors.E("payload[file]", http.StatusRequestEntityTooLarge)
				}
				return errors.E("payload[file]", http.StatusBadRequest)
			}

			continue
		}
	}
	return nil
}
