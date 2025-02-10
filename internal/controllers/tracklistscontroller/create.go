package tracklistscontroller

import (
	"context"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackliststore"
)

// CreateRequest defines the data to read from the HTTP request.
type CreateRequest struct {
	Name    string     `json:"name"`
	Date    string     `json:"date"`
	URL     string     `json:"url"`
	Artwork string     `json:"artwork"`
	Tracks  [][]string `json:"tracks"`
}

// CreateResponse defines the data to write to the HTTP response.
type CreateResponse struct {
	Tracklist *trackliststore.Tracklist `json:"data"`
}

// Create returns an action function that creates a new tracklist.
func Create(tracklistStore *trackliststore.Store) controllers.ActionFunc[CreateRequest, *CreateResponse] {
	return func(ctx context.Context, input CreateRequest) (*CreateResponse, error) {
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

		return &CreateResponse{Tracklist: tracklist}, nil
	}
}
