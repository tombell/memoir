package tracklistscontroller

import (
	"context"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackliststore"
)

// ShowRequest defines the data to read from the HTTP request.
type ShowRequest struct {
	ID string `path:"id"`
}

// ShowResponse defines the data to write to the HTTP response.
type ShowResponse struct {
	Tracklist *trackliststore.Tracklist `json:"data"`
}

// Show returns an action function for getting the track with the given ID.
func Show(tracklistStore *trackliststore.Store) controllers.ActionFunc[ShowRequest, *ShowResponse] {
	return func(ctx context.Context, input ShowRequest) (*ShowResponse, error) {
		tracklist, err := tracklistStore.GetTracklist(ctx, input.ID)
		if err != nil {
			return nil, err
		}

		return &ShowResponse{Tracklist: tracklist}, nil
	}
}
