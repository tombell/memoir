package searchcontroller

import (
	"context"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

// SearchTracksRequest defines the data to read from the HTTP request.
type SearchTracksRequest struct {
	Query   string `query:"q"`
	Page    string `query:"page"`
	PerPage string `query:"per_page"`
}

// SearchTracksResponse defines the data to write to the HTTP response.
type SearchTracksResponse struct {
	Meta   controllers.Meta    `json:"meta"`
	Tracks []*trackstore.Track `json:"data"`
}

// Tracks returns an action function for searching tracks using the track store.
func Tracks(trackStore *trackstore.Store) controllers.ActionFunc[SearchTracksRequest, *SearchTracksResponse] {
	return func(ctx context.Context, input SearchTracksRequest) (*SearchTracksResponse, error) {
		page, perPage, err := controllers.Pagination(input.Page, input.PerPage)
		if err != nil {
			return nil, err
		}

		tracks, total, err := trackStore.SearchTracks(ctx, input.Query, page, perPage)
		if err != nil {
			return nil, err
		}

		return &SearchTracksResponse{Meta: controllers.NewMeta(total, page, perPage), Tracks: tracks}, nil
	}
}
