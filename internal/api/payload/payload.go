package payload

import (
	"mime/multipart"
)

// File is the multipart file and file header read from the form data of an
// HTTP request.
type File struct {
	File   multipart.File
	Header *multipart.FileHeader
}

type ClientReporter interface {
	Message() map[string][]string
	Status() int
}

// StatusCoder is an interface for a type for responses that wish to set a
// specific HTTP status code for the response.
type StatusCoder interface {
	StatusCode() int
}

// ErrorResponse is used for writing errors as JSON to the HTTP response.
type ErrorResponse struct {
	Errors map[string][]string `json:"errors"`

	status int
}

// StatusCode returns the status code for the HTTP response.
func (e ErrorResponse) StatusCode() int {
	return e.status
}
