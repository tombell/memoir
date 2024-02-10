package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/tombell/memoir/internal/datastore"
)

func (s *Services) GetTrack(id string) (*Track, error) {
	row, err := s.DataStore.GetTrack(context.Background(), id)
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

func (s *Services) GetMostPlayedTracks(limit int32) ([]*Track, error) {
	rows, err := s.DataStore.GetMostPlayedTracks(context.Background(), limit)
	if err != nil {
		return nil, fmt.Errorf("find most played tracks failed: %w", err)
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

	return tracks, nil
}

func (s *Services) SearchTracks(query string, limit int32) ([]*Track, error) {
	rows, err := s.DataStore.GetTracksByQuery(context.Background(), datastore.GetTracksByQueryParams{
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
