package tracklistscontroller

import (
	"context"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackliststore"
)

type CreateTracklistRequest struct {
	Name    string     `json:"name"`
	Date    string     `json:"date"`
	URL     string     `json:"url"`
	Artwork string     `json:"artwork"`
	Tracks  [][]string `json:"tracks"`
}

type CreateTracklistsResponse struct {
	Tracklist *trackliststore.Tracklist `json:"data"`
}

func Create(tracklistStore *trackliststore.Store) controllers.ActionFunc[CreateTracklistRequest, *CreateTracklistsResponse] {
	return func(ctx context.Context, input CreateTracklistRequest) (*CreateTracklistsResponse, error) {
		params := &trackliststore.AddTracklistParams{
			Name:    input.Name,
			Date:    input.Date,
			URL:     input.URL,
			Artwork: input.Artwork,
			Tracks:  input.Tracks,
		}

		tracklist, err := tracklistStore.AddTracklist(ctx, params)
		if err != nil {
			return nil, err
		}

		return &CreateTracklistsResponse{Tracklist: tracklist}, nil
	}
}
