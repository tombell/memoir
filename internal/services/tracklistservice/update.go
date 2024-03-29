package tracklistservice

import (
	"context"
	"time"

	"github.com/tombell/memoir/internal/services"
	"github.com/tombell/memoir/internal/stores/trackliststore"
)

type UpdateTracklistRequest struct {
	ID string `path:"id"`

	Name string    `json:"name"`
	Date time.Time `json:"date"`
	URL  string    `json:"url"`
}

type UpdateTracklistsResponse struct {
	Tracklist *trackliststore.Tracklist `json:"tracklist"`
}

func Update(tracklistStore *trackliststore.Store) services.ServiceFunc[UpdateTracklistRequest, *UpdateTracklistsResponse] {
	return func(ctx context.Context, input UpdateTracklistRequest) (*UpdateTracklistsResponse, error) {
		params := &trackliststore.UpdateTracklistParams{
			Name: input.Name,
			Date: input.Date,
			URL:  input.URL,
		}

		tracklist, err := tracklistStore.UpdateTracklist(ctx, input.ID, params)
		if err != nil {
			return nil, err
		}

		return &UpdateTracklistsResponse{Tracklist: tracklist}, nil
	}
}
