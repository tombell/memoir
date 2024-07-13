package trackservice

import (
	"context"

	"github.com/tombell/memoir/internal/services"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

const maxSearchResults = 10

type SearchRequest struct {
	Query string `query:"q"`
}

type SearchResponse struct {
	Tracks []*trackstore.Track `json:"data"`
}

func Search(trackStore *trackstore.Store) services.ServiceFunc[SearchRequest, *SearchResponse] {
	return func(ctx context.Context, input SearchRequest) (*SearchResponse, error) {
		tracks, err := trackStore.SearchTracks(ctx, input.Query, maxSearchResults)
		if err != nil {
			return nil, err
		}

		return &SearchResponse{Tracks: tracks}, nil
	}
}
