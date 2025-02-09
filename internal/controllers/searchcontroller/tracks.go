package searchcontroller

import (
	"context"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

type SearchRequest struct {
	Query   string `query:"q"`
	Page    string `query:"page"`
	PerPage string `query:"per_page"`
}

type SearchResponse struct {
	Tracks []*trackstore.Track `json:"data"`
}

func Tracks(trackStore *trackstore.Store) controllers.ActionFunc[SearchRequest, *SearchResponse] {
	return func(ctx context.Context, input SearchRequest) (*SearchResponse, error) {
		// TODO: implement pagination
		// page, err := controllers.IntQueryParam(input.Page, 1)
		// if err != nil {
		// 	return nil, err
		// }

		perPage, err := controllers.IntQueryParam(input.PerPage, 10)
		if err != nil {
			return nil, err
		}

		tracks, err := trackStore.SearchTracks(ctx, input.Query, perPage)
		if err != nil {
			return nil, err
		}

		return &SearchResponse{Tracks: tracks}, nil
	}
}
