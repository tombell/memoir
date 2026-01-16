package errors

import (
	goerr "errors"
	"fmt"
	"maps"
	"net/http"
	"strings"
)

// M is a type alias for a map of string to string slice.
type M map[string][]string

// Op is the operation being performed, usually the package and method name.
type Op string

// Strf is a wrapper around the fmt.Errorf function.
func Strf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// Error is an error that occured during execution. Includes any error
// messages and HTTP status code for the response. Implements the Error
// interface.
type Error struct {
	op       Op
	err      error
	messages M
	status   int
}

// E creates a new Error with the given op and arguments. The arguments can be
// 1: an error, 2: a M instance, 3: an int status code.
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
			maps.Copy(e.messages, arg)
		case int:
			e.status = arg
		}
	}

	return e
}

// Error returns a string containing all the error information. This should not
// be passed down to any HTTP responses.
func (e *Error) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s: ", string(e.op))

	if e.err != nil {
		b.WriteString(e.err.Error())
	} else {
		b.WriteString("error")
	}

	return b.String()
}

// Unwrap returns the current error's underlying error if there is one.
func (e *Error) Unrap() error {
	return e.err
}

// Message returns a map of the error messages to be returned in the HTTP
// response.
func (e *Error) Message() M {
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

// Status returns the HTTP status code for the error.
func (e *Error) Status() int {
	if e.status >= http.StatusBadRequest {
		return e.status
	}

	return http.StatusInternalServerError
}

// Is is a wrapper around the native errors.Is function.
func Is(err error, target error) bool {
	return goerr.Is(err, target)
}

// As is a wrapper around the native errors.As function.
func As(err error, target any) bool {
	return goerr.As(err, target)
}
