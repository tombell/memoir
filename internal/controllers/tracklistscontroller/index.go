package tracklistscontroller

import (
	"context"

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

func Index(tracklistStore *trackliststore.Store) controllers.ActionFunc[TracklistsRequest, *TracklistsResponse] {
	return func(ctx context.Context, input TracklistsRequest) (*TracklistsResponse, error) {
		page, err := controllers.ParamAsInt(input.Page, 1)
		if err != nil {
			return nil, err
		}

		perPage, err := controllers.ParamAsInt(input.PerPage, 10)
		if err != nil {
			return nil, err
		}

		tracklists, total, err := tracklistStore.GetTracklists(ctx, page, perPage)
		if err != nil {
			return nil, err
		}

		resp := &TracklistsResponse{
			Meta:       controllers.NewMeta(total, page, perPage),
			Tracklists: tracklists,
		}

		return resp, nil
	}
}
