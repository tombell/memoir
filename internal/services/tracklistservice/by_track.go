package tracklistservice

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tombell/memoir/internal/services"
	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

type ByTrackRequest struct {
	Page string `query:"page"`

	ID string `path:"id"`
}

type ByTrackResponse struct {
	CurrentPage string `header:"Current-Page" json:"-"`
	TotalPages  string `header:"Total-Pages" json:"-"`

	Tracklists []*trackliststore.Tracklist `json:"tracklists"`
}

func ByTrack(
	trackStore *trackstore.Store,
	tracklistStore *trackliststore.Store,
) services.ServiceFunc[ByTrackRequest, *ByTrackResponse] {
	return func(ctx context.Context, input ByTrackRequest) (*ByTrackResponse, error) {
		page, _ := strconv.Atoi(input.Page)
		if page <= 0 {
			page = 1
		}

		track, err := trackStore.GetTrack(ctx, input.ID)
		if err != nil {
			return nil, err
		}

		tracklists, total, err := tracklistStore.GetTracklistsByTrack(ctx, track.ID, int32(page), tracklistsPerPage)
		if err != nil {
			return nil, err
		}

		resp := &ByTrackResponse{
			CurrentPage: input.Page,
			TotalPages:  fmt.Sprintf("%d", total),
			Tracklists:  tracklists,
		}

		return resp, nil
	}
}
