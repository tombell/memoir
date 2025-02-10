package controllers

import (
	"context"
)

// ActionFunc is a type that defines a function used as a http Handler that
// reads data from an HTTP request, and also writes data to the HTTP response.
type ActionFunc[In, Out any] func(context.Context, In) (Out, error)

// WriteOnlyActionFunc is a type that defines a function used as a http Handler
// that only writes data to the HTTP response.
type WriteOnlyActionFunc[Out any] func(context.Context) (Out, error)
