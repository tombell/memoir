package tracklistservice

import (
	"context"
	"strconv"

	"github.com/tombell/memoir/internal/services"
	"github.com/tombell/memoir/internal/stores/trackliststore"
)

type TracklistsRequest struct {
	Page string `query:"page"`
}

type TracklistsResponse struct {
	Meta       services.Meta               `json:"meta"`
	Tracklists []*trackliststore.Tracklist `json:"data"`
}

func Index(tracklistStore *trackliststore.Store) services.ServiceFunc[TracklistsRequest, *TracklistsResponse] {
	return func(ctx context.Context, input TracklistsRequest) (*TracklistsResponse, error) {
		page, _ := strconv.Atoi(input.Page)
		if page <= 0 {
			page = 1
		}

		tracklists, total, err := tracklistStore.GetTracklists(ctx, int32(page), tracklistsPerPage)
		if err != nil {
			return nil, err
		}

		resp := &TracklistsResponse{
			Meta:       services.NewMeta(page, total),
			Tracklists: tracklists,
		}

		return resp, nil
	}
}
