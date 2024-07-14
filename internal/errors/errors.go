package errors

import (
	goerr "errors"
	"fmt"
	"net/http"
	"strings"
)

type M map[string][]string

type Op string

func Strf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

type Error struct {
	op       Op
	err      error
	messages M
	status   int
}

func E(op Op, args ...any) error {
	e := &Error{
		op:       op,
		status:   http.StatusInternalServerError,
		messages: M{},
	}

	for _, arg := range args {
		switch arg := arg.(type) {
		case error:
			e.err = arg
		case M:
			for k, v := range arg {
				e.messages[k] = v
			}
		case int:
			e.status = arg
		}
	}

	return e
}

func (e *Error) Error() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s: ", string(e.op)))

	if e.err != nil {
		b.WriteString(e.err.Error())
	} else {
		b.WriteString("error")
	}

	return b.String()
}

func (e *Error) Unrap() error {
	return e.err
}

func (e *Error) Message() map[string][]string {
	if len(e.messages) == 0 {
		switch e.status {
		case http.StatusBadRequest:
			return M{"message": []string{"bad request"}}
		case http.StatusUnauthorized:
			return M{"message": []string{"unauthorized"}}
		case http.StatusForbidden:
			return M{"message": []string{"forbidden"}}
		case http.StatusNotFound:
			return M{"message": []string{"not found"}}
		default:
			return M{"message": []string{"something went wrong"}}
		}
	}

	return e.messages
}

func (e *Error) Status() int {
	if e.status >= http.StatusBadRequest {
		return e.status
	}

	return http.StatusInternalServerError
}

func Is(err error, target error) bool {
	return goerr.Is(err, target)
}

func As(err error, target any) bool {
	return goerr.As(err, target)
}
