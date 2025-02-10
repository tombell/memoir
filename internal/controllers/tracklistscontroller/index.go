package tracklistscontroller

import (
	"context"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackliststore"
)

// IndexRequest defines the data to read from the HTTP request.
type IndexRequest struct {
	Page    string `query:"page"`
	PerPage string `query:"per_page"`
}

// IndexResponse defines the data to write to the HTTP response.
type IndexResponse struct {
	Meta       controllers.Meta            `json:"meta"`
	Tracklists []*trackliststore.Tracklist `json:"data"`
}

// Index returns an action function for getting a list of tracklists.
func Index(tracklistStore *trackliststore.Store) controllers.ActionFunc[IndexRequest, *IndexResponse] {
	return func(ctx context.Context, input IndexRequest) (*IndexResponse, error) {
		page, err := controllers.ParamAsInt(input.Page, 1)
		if err != nil {
			return nil, err
		}

		perPage, err := controllers.ParamAsInt(input.PerPage, 10)
		if err != nil {
			return nil, err
		}

		tracklists, total, err := tracklistStore.GetTracklists(ctx, page, perPage)
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
