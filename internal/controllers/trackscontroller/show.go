package trackscontroller

import (
	"context"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

// ShowRequest defines the data to read from the HTTP request.
type ShowRequest struct {
	ID string `path:"id"`
}

// ShowResponse defines the data to write to the HTTP response.
type ShowResponse struct {
	Track *trackstore.Track `json:"data"`
}

// Show returns an action function for getting the tracklist with the given ID.
func Show(trackStore *trackstore.Store) controllers.ActionFunc[ShowRequest, *ShowResponse] {
	return func(ctx context.Context, input ShowRequest) (*ShowResponse, error) {
		track, err := trackStore.GetTrack(ctx, input.ID)
		if err != nil {
			return nil, err
		}

		return &ShowResponse{Track: track}, nil
	}
}
