package datastore

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/tombell/memoir/internal/datastore/queries"
)

type Track struct {
	ID                string
	Artist            string
	ArtistHighlighted string
	NameHighlighted   string
	Name              string
	Genre             string
	BPM               float64
	Key               string
	Created           time.Time
	Updated           time.Time

	Played int
}

func (s *Store) AddTrack(tx *sql.Tx, track *Track) error {
	if _, err := tx.Exec(
		queries.AddTrack,
		track.ID,
		track.Artist,
		track.Name,
		track.Genre,
		track.BPM,
		track.Key,
		track.Created,
		track.Updated,
	); err != nil {
		return fmt.Errorf("tx exec failed: %w", err)
	}

	return nil
}

func (s *Store) FindTrack(id string) (*Track, error) {
	var track Track

	switch err := s.QueryRowx(queries.FindTrackByID, id).StructScan(&track); err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &track, nil
	default:
		return nil, fmt.Errorf("query row failed: %w", err)
	}
}

func (s *Store) FindTrackByArtistAndName(artist, name string) (*Track, error) {
	var track Track

	switch err := s.QueryRowx(queries.FindTrackByArtistAndName, artist, name).StructScan(&track); err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &track, nil
	default:
		return nil, fmt.Errorf("query row failed: %w", err)
	}
}

func (s *Store) FindMostPlayedTracks(limit int) ([]*Track, error) {
	rows, err := s.Queryx(queries.FindMostPlayedTracks, limit)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}

	var tracks []*Track

	for rows.Next() {
		var track Track

		if err := rows.StructScan(&track); err != nil {
			return nil, fmt.Errorf("rows struct scan failed: %w", err)
		}

		tracks = append(tracks, &track)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows next failed: %w", err)
	}

	return tracks, nil
}

func (s *Store) FindTracksByQuery(query string, limit int) ([]*Track, error) {
	rows, err := s.Queryx(queries.FindTracksByQuery, query, limit)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}

	var tracks []*Track

	for rows.Next() {
		var track Track

		if err := rows.StructScan(&track); err != nil {
			return nil, fmt.Errorf("rows struct scan failed: %w", err)
		}

		tracks = append(tracks, &track)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows next failed: %w", err)
	}

	return tracks, nil
}
