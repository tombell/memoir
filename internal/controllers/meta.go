package controllers

// Meta contains the pagination information for the returned data in the HTTP
// response.
type Meta struct {
	CurrentPage int64 `json:"current_page"`
	TotalPages  int64 `json:"total_pages"`
	PerPage     int64 `json:"per_page"`
	Total       int64 `json:"total"`
}

// NewMeta returns a new Meta initialised with the current page and total pages
// calculated.
func NewMeta(total, currentPage, perPage int64) Meta {
	var totalPages int64
	if total > 0 && perPage > 0 {
		totalPages = total / perPage
		if total%perPage != 0 {
			totalPages++
		}
	}
	return Meta{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		PerPage:     perPage,
		Total:       total,
	}
}
