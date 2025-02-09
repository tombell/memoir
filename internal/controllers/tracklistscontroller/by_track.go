package tracklistscontroller

import (
	"context"
	"strconv"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

type ByTrackRequest struct {
	Page    string `query:"page"`
	PerPage string `query:"per_page"`

	ID string `path:"id"`
}

type ByTrackResponse struct {
	Meta       controllers.Meta            `json:"meta"`
	Tracklists []*trackliststore.Tracklist `json:"data"`
}

func ByTrack(
	trackStore *trackstore.Store,
	tracklistStore *trackliststore.Store,
) controllers.ServiceFunc[ByTrackRequest, *ByTrackResponse] {
	return func(ctx context.Context, input ByTrackRequest) (*ByTrackResponse, error) {
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

		track, err := trackStore.GetTrack(ctx, input.ID)
		if err != nil {
			return nil, err
		}

		tracklists, total, err := tracklistStore.GetTracklistsByTrack(ctx, track.ID, page, perPage)
		if err != nil {
			return nil, err
		}

		resp := &ByTrackResponse{
			Meta:       controllers.NewMeta(total, page, perPage),
			Tracklists: tracklists,
		}

		return resp, nil
	}
}
