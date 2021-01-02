package datastore

import (
	"database/sql"
	"fmt"
	"time"
)

// Track contains data about a track row in the database.
type Track struct {
	ID      string
	Artist  string
	Name    string
	Genre   string
	BPM     float64
	Key     string
	Created time.Time
	Updated time.Time

	Played int
}

// TrackSearchResult contains data about track search result matching on the
// artist or name.
type TrackSearchResult struct {
	Track

	ArtistHighlighted string
	NameHighlighted   string
}

// AddTrack adds a new track into the database.
func (s *Store) AddTrack(tx *sql.Tx, track *Track) error {
	_, err := tx.Exec(insertTrackSQL,
		track.ID,
		track.Artist,
		track.Name,
		track.Genre,
		track.BPM,
		track.Key,
		track.Created,
		track.Updated)

	if err != nil {
		return fmt.Errorf("tx exec failed: %w", err)
	}

	return nil
}

// GetTrack gets a track with the given ID from the database.
func (s *Store) GetTrack(id string) (*Track, error) {
	var track Track

	err := s.QueryRowx(findTrackByIDSQL, id).StructScan(&track)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("query row failed: %w", err)
	default:
		return &track, nil
	}
}

// FindTrackByArtistAndName finds a track with the given artist and name in the
// database.
func (s *Store) FindTrackByArtistAndName(artist, name string) (*Track, error) {
	var track Track

	err := s.QueryRowx(findTrackByArtistAndNameSQL, artist, name).StructScan(&track)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("query row failed: %w", err)
	default:
		return &track, nil
	}
}

// FindMostPlayedTracks finds the tracks that are most played, limiting it to
// the given count in the database.
func (s *Store) FindMostPlayedTracks(limit int) ([]*Track, error) {
	rows, err := s.Queryx(findMostPlayedTracksSQL, limit)
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
func (s *Store) FindTracksByQuery(query string) ([]*TrackSearchResult, error) {
	rows, err := s.Queryx(findTracksByQuerySQL, query)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}

	var tracks []*TrackSearchResult

	for rows.Next() {
		var track TrackSearchResult

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

// UpdateTracksTSVector updates the tsvector column of all tracks.
func (s *Store) UpdateTracksTSVector(tx *sql.Tx) error {
	if _, err := tx.Exec(updateTracksTSVectorSQL); err != nil {
		return fmt.Errorf("tx exec failed: %w", err)
	}

	return nil
}
