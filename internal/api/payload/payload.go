package payload

import (
	"mime/multipart"
)

type File struct {
	File   multipart.File
	Header *multipart.FileHeader
}

type ClientReporter interface {
	Message() map[string][]string
	Status() int
}

type StatusCoder interface {
	StatusCode() int
}

type ErrorResponse struct {
	Error map[string][]string `json:"errors"`

	status int
}

func (e ErrorResponse) StatusCode() int {
	return e.status
}
