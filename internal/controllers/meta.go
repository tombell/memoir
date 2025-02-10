package controllers

import (
	"math"
)

// Meta contains the pagination information for the returned data in the HTTP
// response.
type Meta struct {
	CurrentPage int64   `json:"current_page"`
	TotalPages  float64 `json:"total_pages"`
}

// NewMeta returns a new Meta initialised with the current page and total pages
// calculated.
func NewMeta(total, currentPage, perPage int64) Meta {
	return Meta{
		CurrentPage: currentPage,
		TotalPages:  math.Ceil(float64(total) / float64(perPage)),
	}
}
