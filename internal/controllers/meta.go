package controllers

import (
	"math"
)

type Meta struct {
	CurrentPage int64   `json:"current_page"`
	TotalPages  float64 `json:"total_pages"`
}

func NewMeta(total, currentPage, perPage int64) Meta {
	return Meta{
		CurrentPage: currentPage,
		TotalPages:  math.Ceil(float64(total) / float64(perPage)),
	}
}
