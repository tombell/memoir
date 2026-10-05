package payload

import (
	"errors"
	"net/http"
	"reflect"
)

// Cleanup closes files read into the input and removes multipart temporary files.
// Call it even when Read returns an error, since the input may be partly decoded.
func Cleanup[T any](r *http.Request, in T) error {
	var errs []error

	value := reflect.ValueOf(in)
	if value.Kind() == reflect.Struct {
		for i := range value.NumField() {
			if _, ok := value.Type().Field(i).Tag.Lookup("file"); !ok {
				continue
			}
			file, ok := value.Field(i).Interface().(*File)
			if ok && file != nil && file.File != nil {
				if err := file.File.Close(); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}

	if r.MultipartForm != nil {
		if err := r.MultipartForm.RemoveAll(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
