package payload

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
