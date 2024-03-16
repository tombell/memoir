package tracklistservice

import (
	"context"

	"github.com/tombell/memoir/internal/services"
	"github.com/tombell/memoir/internal/stores/trackliststore"
)

type TracklistRequest struct {
	ID string `path:"id"`
}

type TracklistResponse struct {
	Tracklist *trackliststore.Tracklist `json:"tracklist"`
}

func Show(tracklistStore *trackliststore.Store) services.ServiceFunc[TracklistRequest, *TracklistResponse] {
	return func(ctx context.Context, input TracklistRequest) (*TracklistResponse, error) {
		tracklist, err := tracklistStore.GetTracklist(ctx, input.ID)
		if err != nil {
			return nil, err
		}

		return &TracklistResponse{Tracklist: tracklist}, nil
	}
}
