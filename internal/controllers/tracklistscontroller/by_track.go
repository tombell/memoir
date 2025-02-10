package tracklistscontroller

import (
	"context"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

// ByTrackRequest defines the data to read from the HTTP request.
type ByTrackRequest struct {
	Page    string `query:"page"`
	PerPage string `query:"per_page"`

	ID string `path:"id"`
}

// ByTrackResponse defines the data to write to the HTTP response.
type ByTrackResponse struct {
	Meta       controllers.Meta            `json:"meta"`
	Tracklists []*trackliststore.Tracklist `json:"data"`
}

// ByTrack returns an action function for getting a list of tracklists that
// contain the track with the given ID.
func ByTrack(
	trackStore *trackstore.Store,
	tracklistStore *trackliststore.Store,
) controllers.ActionFunc[ByTrackRequest, *ByTrackResponse] {
	return func(ctx context.Context, input ByTrackRequest) (*ByTrackResponse, error) {
		page, err := controllers.ParamAsInt(input.Page, 1)
		if err != nil {
			return nil, err
		}

		perPage, err := controllers.ParamAsInt(input.PerPage, 10)
		if err != nil {
			return nil, err
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
