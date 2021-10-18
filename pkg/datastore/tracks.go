package datastore

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/tombell/memoir/pkg/datastore/queries"
)

// Track contains data about a track row in the database.
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

// AddTrack adds a new track into the database.
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

// GetTrack gets a track with the given ID from the database.
func (s *Store) GetTrack(id string) (*Track, error) {
	var track Track

	switch err := s.QueryRowx(queries.GetTrackByID, id).StructScan(&track); err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return nil, fmt.Errorf("query row failed: %w", err)
	default:
		return &track, nil
	}
}

// GetTrackByArtistAndName finds a track with the given artist and name in the
// database.
func (s *Store) GetTrackByArtistAndName(artist, name string) (*Track, error) {
	var track Track

	switch err := s.QueryRowx(queries.GetTrackByArtistAndName, artist, name).StructScan(&track); err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return nil, fmt.Errorf("query row failed: %w", err)
	default:
		return &track, nil
	}
}

// FindMostPlayedTracks finds the tracks that are most played, limiting it to
// the given count in the database.
func (s *Store) FindMostPlayedTracks(limit int) ([]*Track, error) {
	rows, err := s.Queryx(queries.GetMostPlayedTracks, limit)
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

// FindTracksByQuery finds the tracks that have artists or names matching the
// given query in the database.
func (s *Store) FindTracksByQuery(query string) ([]*Track, error) {
	rows, err := s.Queryx(queries.GetTracksByQuery, query)
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
