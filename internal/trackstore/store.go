package trackstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/tombell/memoir/internal/datastore"
)

type Track struct {
	ID     string  `json:"id"`
	Artist string  `json:"artist"`
	Name   string  `json:"name"`
	Genre  string  `json:"genre"`
	BPM    float64 `json:"bpm"`
	Key    string  `json:"key"`

	Created time.Time `json:"-"`
	Updated time.Time `json:"-"`

	Played int64 `json:"played,omitempty"`

	ArtistHighlighted string `json:"artistHighlighted,omitempty"`
	NameHighlighted   string `json:"nameHighlighted,omitempty"`
}

type Store struct {
	dataStore *datastore.Store
}

func New(store *datastore.Store) *Store {
	return &Store{dataStore: store}
}

func (s *Store) GetTrack(id string) (*Track, error) {
	row, err := s.dataStore.GetTrack(context.Background(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get track failed: %w", err)
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

func (s *Store) GetMostPlayedTracks(limit int32) ([]*Track, error) {
	rows, err := s.dataStore.GetMostPlayedTracks(context.Background(), limit)
	if err != nil {
		return nil, fmt.Errorf("find most played tracks failed: %w", err)
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

func (s *Store) SearchTracks(query string, limit int32) ([]*Track, error) {
	rows, err := s.dataStore.GetTracksByQuery(context.Background(), datastore.GetTracksByQueryParams{
		Query:    query,
		RowLimit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("find tracks by query failed: %w", err)
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
