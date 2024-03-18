package services

import "context"

type ServiceFunc[In, Out any] func(context.Context, In) (Out, error)

type WriteOnlyServiceFunc[Out any] func(context.Context) (Out, error)

type Meta struct {
	CurrentPage int   `json:"current_page"`
	TotalPages  int64 `json:"total_pages"`
}
