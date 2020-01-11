package datastore

import (
	"database/sql"
	"fmt"
	"time"
)

// Tracklist contains data about a tracklist row in the database.
type Tracklist struct {
	ID      string
	Name    string
	Artwork string
	URL     string
	Date    time.Time
	Created time.Time
	Updated time.Time

	TrackCount int
	Tracks     []*Track
}

// AddTracklist adds a new tracklist into the database.
func (s *Store) AddTracklist(tx *sql.Tx, tracklist *Tracklist) error {
	_, err := tx.Exec(insertTracklistSQL,
		tracklist.ID,
		tracklist.Name,
		tracklist.URL,
		tracklist.Artwork,
		tracklist.Date,
		tracklist.Created,
		tracklist.Updated)

	if err != nil {
		return fmt.Errorf("tx exec failed: %w", err)
	}

	return nil
}

// RemoveTracklist removes a tracklist from the database.
func (s *Store) RemoveTracklist(tx *sql.Tx, id string) error {
	if _, err := tx.Exec(deleteTracklistSQL, id); err != nil {
		return fmt.Errorf("tx exec failed: %w", err)
	}

	return nil
}

// AddArtworkToTracklist adds an artwork file key to the given tracklist in the
// database.
func (s *Store) AddArtworkToTracklist(tx *sql.Tx, id, artwork string) error {
	if _, err := tx.Exec(addArtworkToTracklistSQL, artwork, id); err != nil {
		return fmt.Errorf("tx exec failed: %w", err)
	}

	return nil
}

// GetTracklists gets all tracklists.
func (s *Store) GetTracklists() ([]*Tracklist, error) {
	rows, err := s.Queryx(getTracklistsSQL)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}

	var tracklists []*Tracklist

	for rows.Next() {
		var tracklist Tracklist

		if err := rows.StructScan(&tracklist); err != nil {
			return nil, fmt.Errorf("rows scan failed: %w", err)
		}

		tracklists = append(tracklists, &tracklist)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows next failed: %w", err)
	}

	return tracklists, nil
}

// GetTracklist gets a tracklist with the given ID from the database.
func (s *Store) GetTracklist(id string) (*Tracklist, error) {
	var tracklist Tracklist

	err := s.QueryRowx(findTracklistByIDSQL, id).StructScan(&tracklist)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("db query row failed: %w", err)
	default:
		return &tracklist, nil
	}
}

// GetTracklistWithTracks gets a tracklist with the given ID, and associated
// tracks from the database.
func (s *Store) GetTracklistWithTracks(id string) (*Tracklist, error) {
	rows, err := s.Queryx(findTracklistWithTracksByIDSQL, id)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}

	var tracklist Tracklist

	for rows.Next() {
		var track Track

		err := rows.Scan(
			&tracklist.ID,
			&tracklist.Name,
			&tracklist.Date,
			&tracklist.Created,
			&tracklist.Updated,
			&tracklist.Artwork,
			&track.ID,
			&track.Artist,
			&track.Name,
			&track.Genre,
			&track.BPM,
			&track.Key,
			&track.Created,
			&track.Updated)

		if err != nil {
			return nil, fmt.Errorf("rows scan failed: %w", err)
		}

		tracklist.Tracks = append(tracklist.Tracks, &track)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows next failed: %w", err)
	}

	if tracklist.ID == "" {
		return nil, nil
	}

	return &tracklist, nil
}

// FindTracklistByName finds a tracklist with the given name in the database.
func (s *Store) FindTracklistByName(name string) (*Tracklist, error) {
	var tracklist Tracklist

	err := s.QueryRowx(findTracklistByNameSQL, name).StructScan(&tracklist)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("db query row failed: %w", err)
	default:
		return &tracklist, nil
	}
}

// FindTracklistWithTracksByName find a tracklist with the given name, and
// associated tracks in the database.
func (s *Store) FindTracklistWithTracksByName(name string) (*Tracklist, error) {
	rows, err := s.Queryx(findTracklistWithTracksByNameSQL, name)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}

	var tracklist Tracklist

	for rows.Next() {
		var track Track

		err := rows.Scan(
			&tracklist.ID,
			&tracklist.Name,
			&tracklist.Date,
			&tracklist.Created,
			&tracklist.Updated,
			&tracklist.Artwork,
			&track.ID,
			&track.Artist,
			&track.Name,
			&track.Genre,
			&track.BPM,
			&track.Key,
			&track.Created,
			&track.Updated)

		if err != nil {
			return nil, fmt.Errorf("rows scan failed: %w", err)
		}

		tracklist.Tracks = append(tracklist.Tracks, &track)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows next failed: %w", err)
	}

	if tracklist.Name == "" {
		return nil, fmt.Errorf("could not find tracklist: %q", name)
	}

	return &tracklist, nil
}

// FindTracklistsByTrackID finds all tracklists that contain the given track
// in the database.
func (s *Store) FindTracklistsByTrackID(id string) ([]*Tracklist, error) {
	rows, err := s.Queryx(findTracklistByTrackIDSQL, id)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}

	var tracklists []*Tracklist

	for rows.Next() {
		var tracklist Tracklist

		if err := rows.StructScan(&tracklist); err != nil {
			return nil, fmt.Errorf("rows scan failed: %w", err)
		}

		tracklists = append(tracklists, &tracklist)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows next failed: %w", err)
	}

	return tracklists, nil
}
