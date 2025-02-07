package trackscontroller

import (
	"context"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

const maxMostPlayedTracks = 10

type MostPlayedResponse struct {
	Tracks []*trackstore.Track `json:"data"`
}

func MostPlayed(trackStore *trackstore.Store) controllers.WriteOnlyServiceFunc[*MostPlayedResponse] {
	return func(ctx context.Context) (*MostPlayedResponse, error) {
		tracks, err := trackStore.GetMostPlayedTracks(ctx, maxMostPlayedTracks)
		if err != nil {
			return nil, err
		}

		return &MostPlayedResponse{Tracks: tracks}, nil
	}
}
