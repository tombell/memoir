package trackscontroller

import (
	"context"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

// MostPlayedRequest defines the data to read from the HTTP request.
type MostPlayedRequest struct {
	Page    *string `query:"page"`
	PerPage *string `query:"per_page"`
}

// MostPlayedResponse defines the data to write from the HTTP response.
type MostPlayedResponse struct {
	Tracks []*trackstore.Track `json:"data"`
}

// MostPlayed returns an action function that lists the tracks that are
// contained in the most tracklists.
func MostPlayed(trackStore *trackstore.Store) controllers.ActionFunc[MostPlayedRequest, *MostPlayedResponse] {
	return func(ctx context.Context, input MostPlayedRequest) (*MostPlayedResponse, error) {
		// TODO: implement pagination
		_, perPage, err := controllers.PaginationParams(input.Page, input.PerPage)
		if err != nil {
			return nil, err
		}

		tracks, err := trackStore.GetMostPlayedTracks(ctx, perPage)
		if err != nil {
			return nil, err
		}

		return &MostPlayedResponse{Tracks: tracks}, nil
	}
}
