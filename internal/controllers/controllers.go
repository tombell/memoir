package controllers

import (
	"context"
)

type ServiceFunc[In, Out any] func(context.Context, In) (Out, error)

type WriteOnlyServiceFunc[Out any] func(context.Context) (Out, error)
