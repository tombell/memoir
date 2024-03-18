package payload

import (
	"mime/multipart"
)

type File struct {
	File   multipart.File
	Header *multipart.FileHeader
}

type clientReporter interface {
	Message() map[string]string
	StatusCode() int
}

type statusCoder interface {
	StatusCode() int
}

type errorResponse struct {
	Error map[string]string `json:"error"`

	status int
}

func (e errorResponse) StatusCode() int {
	return e.status
}
