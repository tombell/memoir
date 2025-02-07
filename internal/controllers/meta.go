package controllers

import (
	"math"
)

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
