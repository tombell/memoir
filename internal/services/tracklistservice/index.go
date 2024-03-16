package tracklistservice

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tombell/memoir/internal/services"
	"github.com/tombell/memoir/internal/stores/trackliststore"
)

type TracklistsRequest struct {
	Page string `query:"page"`
}

type TracklistsResponse struct {
	CurrentPage string `header:"Current-Page" json:"-"`
	TotalPages  string `header:"Total-Pages" json:"-"`

	Tracklists []*trackliststore.Tracklist `json:"tracklists"`
}

func Index(tracklistStore *trackliststore.Store) services.ServiceFunc[TracklistsRequest, *TracklistsResponse] {
	return func(ctx context.Context, input TracklistsRequest) (*TracklistsResponse, error) {
		page, _ := strconv.Atoi(input.Page)
		if page <= 0 {
			page = 1
		}

		tracklists, total, err := tracklistStore.GetTracklists(ctx, int32(page), 20)
		if err != nil {
			return nil, err
		}

		resp := &TracklistsResponse{
			CurrentPage: input.Page,
			TotalPages:  fmt.Sprintf("%d", total),
			Tracklists:  tracklists,
		}

		return resp, nil
	}
}
