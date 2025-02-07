package tracklistscontroller

import (
	"context"
	"strconv"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackliststore"
)

type TracklistsRequest struct {
	Page string `query:"page"`
}

type TracklistsResponse struct {
	Meta       controllers.Meta            `json:"meta"`
	Tracklists []*trackliststore.Tracklist `json:"data"`
}

func Index(tracklistStore *trackliststore.Store) controllers.ServiceFunc[TracklistsRequest, *TracklistsResponse] {
	return func(ctx context.Context, input TracklistsRequest) (*TracklistsResponse, error) {
		page, _ := strconv.Atoi(input.Page)
		if page <= 0 {
			page = 1
		}

		// TODO: move tracklistsPerPage to incoming query string param, with default?
		tracklists, total, err := tracklistStore.GetTracklists(ctx, int32(page), 10)
		if err != nil {
			return nil, err
		}

		resp := &TracklistsResponse{
			Meta:       controllers.NewMeta(page, total),
			Tracklists: tracklists,
		}

		return resp, nil
	}
}
