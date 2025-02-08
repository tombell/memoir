package tracklistscontroller

import (
	"context"
	"strconv"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackliststore"
)

type TracklistsRequest struct {
	Page    string `query:"page"`
	PerPage string `query:"per_page"`
}

type TracklistsResponse struct {
	Meta       controllers.Meta            `json:"meta"`
	Tracklists []*trackliststore.Tracklist `json:"data"`
}

func Index(tracklistStore *trackliststore.Store) controllers.ServiceFunc[TracklistsRequest, *TracklistsResponse] {
	return func(ctx context.Context, input TracklistsRequest) (*TracklistsResponse, error) {
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

		tracklists, total, err := tracklistStore.GetTracklists(ctx, page, perPage)
		if err != nil {
			return nil, err
		}

		resp := &TracklistsResponse{
			Meta:       controllers.NewMeta(page, total, perPage),
			Tracklists: tracklists,
		}

		return resp, nil
	}
}
