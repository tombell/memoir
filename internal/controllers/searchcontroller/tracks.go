package searchcontroller

import (
	"context"
	"strconv"

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
		if len(input.Page) == 0 {
			input.Page = "1"
		}

		page, err := strconv.ParseInt(input.Page, 10, 64)
		if err != nil {
			return nil, err
		}
		if page <= 0 {
			page = 1
		}

		if len(input.PerPage) == 0 {
			input.PerPage = "10"
		}

		perPage, err := strconv.ParseInt(input.PerPage, 10, 64)
		if err != nil {
			return nil, err
		}
		if perPage <= 0 {
			perPage = 10
		}

		tracks, err := trackStore.SearchTracks(ctx, input.Query, perPage)
		if err != nil {
			return nil, err
		}

		return &SearchResponse{Tracks: tracks}, nil
	}
}
