package trackscontroller

import (
	"context"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

type MostPlayedRequest struct {
	Page    string `query:"page"`
	PerPage string `query:"per_page"`
}

type MostPlayedResponse struct {
	Tracks []*trackstore.Track `json:"data"`
}

func MostPlayed(trackStore *trackstore.Store) controllers.ActionFunc[MostPlayedRequest, *MostPlayedResponse] {
	return func(ctx context.Context, input MostPlayedRequest) (*MostPlayedResponse, error) {
		// TODO: implement pagination
		// page, err := controllers.IntQueryParam(input.Page, 1)
		// if err != nil {
		// 	return nil, err
		// }

		perPage, err := controllers.ParamAsInt(input.PerPage, 10)
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
