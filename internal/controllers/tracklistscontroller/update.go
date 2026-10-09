package tracklistscontroller

import (
	"context"
	"net/http"

	"github.com/tombell/memoir/internal/auth"
	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/errors"
	"github.com/tombell/memoir/internal/stores/trackliststore"
)

// UpdateRequest defines the data to read from the HTTP request.
type UpdateRequest struct {
	ID string `path:"id"`

	Name string `json:"name"`
	Date string `json:"date"`
	URL  string `json:"url"`
}

// UpdateResponse defines the data to write to the HTTP response.
type UpdateResponse struct {
	Tracklist *trackliststore.Tracklist `json:"data"`
}

// Update returns an action function that updates a tracklist with the given ID.
func Update(tracklistStore *trackliststore.Store) controllers.ActionFunc[UpdateRequest, *UpdateResponse] {
	return func(ctx context.Context, input UpdateRequest) (*UpdateResponse, error) {
		user, ok := auth.UserFromContext(ctx)
		if !ok {
			return nil, errors.E("tracklists[update]", http.StatusUnauthorized)
		}
		params := &trackliststore.UpdateTracklistParams{
			Name: input.Name,
			Date: input.Date,
			URL:  input.URL,
		}

		tracklist, err := tracklistStore.UpdateTracklist(ctx, input.ID, user.ID, params)
		if err != nil {
			return nil, err
		}

		return &UpdateResponse{Tracklist: tracklist}, nil
	}
}
