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

type Store struct {
	dataStore *datastore.Store
}

func New(store *datastore.Store) *Store {
	return &Store{dataStore: store}
}

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

func (s *Store) GetMostPlayedTracks(ctx context.Context, limit int32) ([]*Track, error) {
	op := errors.Op("trackstore[get-most-played-tracks]")

	rows, err := s.dataStore.GetMostPlayedTracks(ctx, limit)
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

func (s *Store) SearchTracks(ctx context.Context, query string, limit int32) ([]*Track, error) {
	op := errors.Op("trackstore[search-tracks]")

	rows, err := s.dataStore.GetTracksByQuery(ctx, db.GetTracksByQueryParams{
		Query:    query,
		RowLimit: limit,
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
