package tracklistscontroller

import (
	"context"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackliststore"
)

// UpdateRequest defines the data to read from the HTTP request.
type UpdateRequest struct {
	ID string `path:"id"`

	Name    trackliststore.PatchField[string] `json:"name"`
	Date    trackliststore.PatchField[string] `json:"date"`
	URL     trackliststore.PatchField[string] `json:"url"`
	Artwork trackliststore.PatchField[string] `json:"artwork"`
}

// UpdateResponse defines the data to write to the HTTP response.
type UpdateResponse struct {
	Tracklist *trackliststore.Tracklist `json:"data"`
}

// Update returns an action function that updates a tracklist with the given ID.
func Update(tracklistStore *trackliststore.Store) controllers.ActionFunc[UpdateRequest, *UpdateResponse] {
	return func(ctx context.Context, input UpdateRequest) (*UpdateResponse, error) {
		params := &trackliststore.UpdateTracklistParams{
			Name:    input.Name,
			Date:    input.Date,
			URL:     input.URL,
			Artwork: input.Artwork,
		}

		tracklist, err := tracklistStore.UpdateTracklist(ctx, input.ID, params)
		if err != nil {
			return nil, err
		}

		return &UpdateResponse{Tracklist: tracklist}, nil
	}
}
