package tracklistscontroller

import (
	"context"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

// IndexRequest defines the data to read from the HTTP request.
type IndexRequest struct {
	// Pagination
	Page    string `query:"page"`
	PerPage string `query:"per_page"`

	// Filters
	TrackID string `query:"track_id"`
	UserID  string `query:"user_id"`
}

// IndexResponse defines the data to write to the HTTP response.
type IndexResponse struct {
	Meta       controllers.Meta            `json:"meta"`
	Tracklists []*trackliststore.Tracklist `json:"data"`
}

// Index returns an action function for getting a list of tracklists.
func Index(
	trackStore *trackstore.Store,
	tracklistStore *trackliststore.Store,
) controllers.ActionFunc[IndexRequest, *IndexResponse] {
	return func(ctx context.Context, input IndexRequest) (*IndexResponse, error) {
		page, err := controllers.ParamAsInt(input.Page, 1)
		if err != nil {
			return nil, err
		}

		perPage, err := controllers.ParamAsInt(input.PerPage, 10)
		if err != nil {
			return nil, err
		}

		if perPage > 100 {
			perPage = 100
		}

		if input.TrackID != "" {
			track, err := trackStore.GetTrack(ctx, input.TrackID)
			if err != nil {
				return nil, err
			}

			input.TrackID = track.ID
		}
		tracklists, total, err := tracklistStore.GetTracklists(ctx, input.UserID, input.TrackID, page, perPage)
		if err != nil {
			return nil, err
		}

		resp := &IndexResponse{
			Meta:       controllers.NewMeta(total, page, perPage),
			Tracklists: tracklists,
		}

		return resp, nil
	}
}
