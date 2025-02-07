package trackscontroller

import (
	"context"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

type TrackRequest struct {
	ID string `path:"id"`
}

type TrackResponse struct {
	Track *trackstore.Track `json:"data"`
}

func Show(trackStore *trackstore.Store) controllers.ServiceFunc[TrackRequest, *TrackResponse] {
	return func(ctx context.Context, input TrackRequest) (*TrackResponse, error) {
		track, err := trackStore.GetTrack(ctx, input.ID)
		if err != nil {
			return nil, err
		}

		return &TrackResponse{Track: track}, nil
	}
}
