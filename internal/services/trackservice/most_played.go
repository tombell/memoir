package trackservice

import (
	"context"

	"github.com/tombell/memoir/internal/services"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

const maxMostPlayedTracks = 10

type MostPlayedResponse struct {
	Tracks []*trackstore.Track `json:"tracks"`
}

func MostPlayed(trackStore *trackstore.Store) services.WriteOnlyServiceFunc[*MostPlayedResponse] {
	return func(ctx context.Context) (*MostPlayedResponse, error) {
		tracks, err := trackStore.GetMostPlayedTracks(ctx, maxMostPlayedTracks)
		if err != nil {
			return nil, err
		}

		return &MostPlayedResponse{Tracks: tracks}, nil
	}
}
