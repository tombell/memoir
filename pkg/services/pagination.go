package services

import (
	"math"
)

// PagedTracklists contains a paginated list of tracklists, indicating if
// another page is available.
type PagedTracklists struct {
	Tracklists []*Tracklist `json:"tracklists"`
	HasMore    bool         `json:"hasMore"`
}

// NewPagedTracklists returns a new PagedTracklists for the given list of
// tracklists based on the given page and per page amount.
func NewPagedTracklists(tracklists []*Tracklist, page, perPage int) *PagedTracklists {
	total := len(tracklists)
	pages := math.Ceil(float64(total) / float64(perPage))
	offset := (page - 1) * perPage
	count := offset + perPage

	if count > total {
		count = total
	}

	return &PagedTracklists{
		Tracklists: tracklists[offset:count],
		HasMore:    page < int(pages),
	}
}

// PagedTracks contains a paginated list of tracks, indicating if another page
// is available.
type PagedTracks struct {
	Tracks  []*Track `json:"tracks"`
	HasMore bool     `json:"hasHore"`
}

// NewPagedTracks returns a new PagedTracks for the given list of
// tracks based on the given page and per page amount.
func NewPagedTracks(tracks []*Track, page, perPage int) *PagedTracks {
	total := len(tracks)
	pages := math.Ceil(float64(total) / float64(perPage))
	offset := (page - 1) * perPage
	count := offset + perPage

	if count > total {
		count = total
	}

	return &PagedTracks{
		Tracks:  tracks[offset:count],
		HasMore: page < int(pages),
	}
}
