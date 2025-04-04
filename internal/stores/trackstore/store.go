package trackstore

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	db "github.com/tombell/memoir/internal/database"
	"github.com/tombell/memoir/internal/errors"
	"github.com/tombell/memoir/internal/stores/datastore"
)

// Store is a store for interacting with tracks in the data store.
type Store struct {
	dataStore *datastore.Store
}

// New returns a new store.
func New(store *datastore.Store) *Store {
	return &Store{dataStore: store}
}

// GetTrack returns a track with the given ID.
// If no track exists return a not found error.
func (s *Store) GetTrack(ctx context.Context, id string) (*Track, error) {
	op := errors.Op("trackstore[get-track]")

	_, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.E(op, http.StatusNotFound)
	}

	row, err := s.dataStore.GetTrack(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.E(op, http.StatusNotFound)
		}

		return nil, errors.E(op, errors.Strf("get track failed: %w", err))
	}

	return &Track{
		ID:      row.ID,
		Artist:  row.Artist,
		Name:    row.Name,
		Genre:   row.Genre,
		BPM:     row.BPM,
		Key:     row.Key,
		Created: row.Created,
		Updated: row.Updated,
	}, nil
}

// GetMostPlayedTracks returns a list of the tracks that are contained in the
// most tracklists. The list is limited with the limit argument.
// TODO: properly paginate.
func (s *Store) GetMostPlayedTracks(ctx context.Context, limit int64) ([]*Track, error) {
	op := errors.Op("trackstore[get-most-played-tracks]")

	rows, err := s.dataStore.GetMostPlayedTracks(ctx, int32(limit))
	if err != nil {
		return nil, errors.E(op, errors.Strf("find most played tracks failed: %w", err))
	}

	tracks := make([]*Track, 0, len(rows))

	for _, row := range rows {
		track := &Track{
			ID:      row.Track.ID,
			Name:    row.Track.Name,
			Artist:  row.Track.Artist,
			BPM:     row.Track.BPM,
			Key:     row.Track.Key,
			Genre:   row.Track.Genre,
			Created: row.Track.Created,
			Updated: row.Track.Updated,
			Played:  row.Played,
		}

		tracks = append(tracks, track)
	}

	return tracks, nil
}

// SearchTracks returns a list of tracks that match the full text search
// results. The list is limited with the limit arugment.
// TODO: properly paginate.
func (s *Store) SearchTracks(ctx context.Context, query string, limit int64) ([]*Track, error) {
	op := errors.Op("trackstore[search-tracks]")

	rows, err := s.dataStore.GetTracksByQuery(ctx, db.GetTracksByQueryParams{
		Query:    query,
		RowLimit: int32(limit),
	})
	if err != nil {
		return nil, errors.E(op, errors.Strf("find tracks by query failed: %w", err))
	}

	tracks := make([]*Track, 0, len(rows))

	for _, row := range rows {
		track := &Track{
			ID:                row.ID,
			Name:              row.Name,
			NameHighlighted:   string(row.NameHighlighted),
			Artist:            row.Artist,
			ArtistHighlighted: string(row.ArtistHighlighted),
			BPM:               row.BPM,
			Key:               row.Key,
			Genre:             row.Genre,
			Created:           row.Created,
			Updated:           row.Updated,
		}

		tracks = append(tracks, track)
	}

	return tracks, nil
}
