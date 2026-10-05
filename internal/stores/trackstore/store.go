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
// most distinct tracklists, together with the total number of played tracks.
func (s *Store) GetMostPlayedTracks(ctx context.Context, page, limit int64) ([]*Track, int64, error) {
	op := errors.Op("trackstore[get-most-played-tracks]")

	total, err := s.dataStore.CountMostPlayedTracks(ctx)
	if err != nil {
		return nil, 0, errors.E(op, errors.Strf("count most played tracks failed: %w", err))
	}
	rows, err := s.dataStore.GetMostPlayedTracks(ctx, db.GetMostPlayedTracksParams{
		RowLimit:  int32(limit),
		RowOffset: int32(limit * (page - 1)),
	})
	if err != nil {
		return nil, 0, errors.E(op, errors.Strf("find most played tracks failed: %w", err))
	}

	tracks := make([]*Track, 0, len(rows))

	for _, row := range rows {
		track := &Track{
			ID:      row.ID,
			Name:    row.Name,
			Artist:  row.Artist,
			BPM:     row.BPM,
			Key:     row.Key,
			Genre:   row.Genre,
			Created: row.Created,
			Updated: row.Updated,
			Played:  row.Played,
		}

		tracks = append(tracks, track)
	}

	return tracks, total, nil
}

// SearchTracks returns a list of tracks that match the full text search
// results, together with the total number of matches before pagination.
func (s *Store) SearchTracks(ctx context.Context, query string, page, limit int64) ([]*Track, int64, error) {
	op := errors.Op("trackstore[search-tracks]")

	total, err := s.dataStore.CountTracksByQuery(ctx, query)
	if err != nil {
		return nil, 0, errors.E(op, errors.Strf("count tracks by query failed: %w", err))
	}
	rows, err := s.dataStore.GetTracksByQuery(ctx, db.GetTracksByQueryParams{
		Query:     query,
		RowLimit:  int32(limit),
		RowOffset: int32(limit * (page - 1)),
	})
	if err != nil {
		return nil, 0, errors.E(op, errors.Strf("find tracks by query failed: %w", err))
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

	return tracks, total, nil
}
