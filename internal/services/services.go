package services

import (
	"context"
	"math"
)

type ServiceFunc[In, Out any] func(context.Context, In) (Out, error)

type WriteOnlyServiceFunc[Out any] func(context.Context) (Out, error)

type Meta struct {
	CurrentPage int   `json:"current_page"`
	TotalPages  int64 `json:"total_pages"`
}

func NewMeta(current int, total int64) Meta {
	return Meta{
		CurrentPage: current,
		TotalPages:  int64(math.Ceil(float64(total) / float64(10))),
	}
}
