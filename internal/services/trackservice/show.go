package trackservice

import (
	"context"

	"github.com/tombell/memoir/internal/services"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

type TrackRequest struct {
	ID string `path:"id"`
}

type TrackResponse struct {
	Track *trackstore.Track `json:"track"`
}

func Show(trackStore *trackstore.Store) services.ServiceFunc[TrackRequest, *TrackResponse] {
	return func(ctx context.Context, input TrackRequest) (*TrackResponse, error) {
		track, err := trackStore.GetTrack(ctx, input.ID)
		if err != nil {
			return nil, err
		}

		return &TrackResponse{Track: track}, nil
	}
}
