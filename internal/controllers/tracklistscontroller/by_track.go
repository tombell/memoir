package tracklistscontroller

import (
	"context"
	"strconv"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

type ByTrackRequest struct {
	Page string `query:"page"`

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
		page, _ := strconv.Atoi(input.Page)
		if page <= 0 {
			page = 1
		}

		track, err := trackStore.GetTrack(ctx, input.ID)
		if err != nil {
			return nil, err
		}

		// TODO: move tracklistsPerPage to incoming query string param, with default?
		tracklists, total, err := tracklistStore.GetTracklistsByTrack(ctx, track.ID, int32(page), 10)
		if err != nil {
			return nil, err
		}

		resp := &ByTrackResponse{
			Meta:       controllers.NewMeta(page, total),
			Tracklists: tracklists,
		}

		return resp, nil
	}
}
