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
	Tracks []*trackstore.Track `json:"data"`
}

// Tracks returns an action function for searching tracks using the track store.
func Tracks(trackStore *trackstore.Store) controllers.ActionFunc[SearchTracksRequest, *SearchTracksResponse] {
	return func(ctx context.Context, input SearchTracksRequest) (*SearchTracksResponse, error) {
		// TODO: implement pagination
		// page, err := controllers.IntQueryParam(input.Page, 1)
		// if err != nil {
		// 	return nil, err
		// }

		perPage, err := controllers.ParamAsInt(input.PerPage, 10)
		if err != nil {
			return nil, err
		}

		tracks, err := trackStore.SearchTracks(ctx, input.Query, perPage)
		if err != nil {
			return nil, err
		}

		return &SearchTracksResponse{Tracks: tracks}, nil
	}
}
