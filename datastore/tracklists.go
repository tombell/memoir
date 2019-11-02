package datastore

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	sqlAddTracklist = `
		INSERT INTO tracklists (
			id,
			name,
			date,
			created,
			updated
		) VALUES ($1, $2, $3, $4, $5)`

	sqlRemoveTracklist = `
		DELETE FROM tracklists
		WHERE id = $1`

	sqlAddArtworkToTracklist = `
		UPDATE tracklists
		SET artwork = $1
		WHERE id = $2`

	sqlGetTracklists = `
		SELECT
			tl.*,
			count(tl.id) as track_count
		FROM tracklists tl
		JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
		GROUP BY tl.id
		ORDER BY tl.date DESC`

	sqlGetTracklistByID = `
		SELECT
			id,
			name,
			date,
			created,
			updated
		FROM tracklists
		WHERE id = $1
		LIMIT 1`

	sqlGetTracklistWithTracksByID = `
		SELECT
			tl.*,
			t.id as track_id,
			t.artist,
			t.name,
			t.genre,
			t.bpm,
			t.key,
			t.created,
			t.updated
		FROM tracklists tl
		JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
		JOIN tracks t ON t.id = tt.track_id
		WHERE tl.id = $1
		ORDER BY tt.track_number ASC`

	sqlFindTracklistByName = `
		SELECT
			id,
			name,
			date,
			created,
			updated
		FROM tracklists
		WHERE name = $1
		LIMIT 1`

	sqlFindTracklistWithTracksByName = `
		SELECT
			tl.*,
			t.id as track_id,
			t.artist,
			t.name,
			t.genre,
			t.bpm,
			t.key,
			t.created,
			t.updated
		FROM tracklists tl
		JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
		JOIN tracks t ON t.id = tt.track_id
		WHERE tl.name = $1
		ORDER BY tt.track_number ASC`

	sqlFindTracklistsByTrackID = `
		SELECT tl.*, (
			SELECT count(id)
			FROM tracklist_tracks
			WHERE tracklist_tracks.tracklist_id = tl.id
		) as track_count
		FROM tracklists tl
		JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
		WHERE tt.track_id = $1
		ORDER BY tl.date DESC`
)

// Tracklist contains data about a tracklist row in the database.
type Tracklist struct {
	ID      string
	Name    string
	Artwork string
	Date    time.Time
	Created time.Time
	Updated time.Time

	TrackCount int
	Tracks     []*Track
}

// AddTracklist adds a new tracklist into the database.
func (s *Store) AddTracklist(tx *sql.Tx, tracklist *Tracklist) error {
	_, err := tx.Exec(sqlAddTracklist,
		tracklist.ID,
		tracklist.Name,
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
	if _, err := tx.Exec(sqlRemoveTracklist, id); err != nil {
		return fmt.Errorf("tx exec failed: %w", err)
	}

	return nil
}

// AddArtworkToTracklist adds an artwork file key to the given tracklist in the
// database.
func (s *Store) AddArtworkToTracklist(tx *sql.Tx, id, artwork string) error {
	if _, err := tx.Exec(sqlAddArtworkToTracklist, artwork, id); err != nil {
		return fmt.Errorf("tx exec failed: %w", err)
	}

	return nil
}

// GetTracklists gets all tracklists.
func (s *Store) GetTracklists() ([]*Tracklist, error) {
	rows, err := s.Queryx(sqlGetTracklists)
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

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

	err := s.QueryRowx(sqlGetTracklistByID, id).StructScan(&tracklist)

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
	rows, err := s.Queryx(sqlGetTracklistWithTracksByID, id)
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

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

	err := s.QueryRowx(sqlFindTracklistByName, name).StructScan(&tracklist)

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
	rows, err := s.Queryx(sqlFindTracklistWithTracksByName, name)
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

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
	rows, err := s.Queryx(sqlFindTracklistsByTrackID, id)
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

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
