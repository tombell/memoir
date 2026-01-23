package tracklistscontroller

import (
	"context"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackliststore"
)

// DeleteRequest defines the data to read from the HTTP request.
type DeleteRequest struct {
	ID string `path:"id"`
}

// DeleteResponse defines the data to write to the HTTP response.
type DeleteResponse struct{}

// Delete returns an action function that deletes a tracklist with the given ID.
func Delete(tracklistStore *trackliststore.Store) controllers.ActionFunc[DeleteRequest, *DeleteResponse] {
	return func(ctx context.Context, input DeleteRequest) (*DeleteResponse, error) {
		err := tracklistStore.DeleteTracklist(ctx, input.ID)
		if err != nil {
			return nil, err
		}

		return &DeleteResponse{}, nil
	}
}
