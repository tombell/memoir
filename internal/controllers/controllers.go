package controllers

import (
	"context"
)

type ActionFunc[In, Out any] func(context.Context, In) (Out, error)

type WriteOnlyActionFunc[Out any] func(context.Context) (Out, error)
