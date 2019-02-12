package datastore

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
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
)

// Tracklist represents a single tracklist row in the database.
type Tracklist struct {
	ID      string
	Name    string
	Date    time.Time
	Created time.Time
	Updated time.Time

	Tracks []*Track
}

// AddTracklist adds a new tracklist into the database.
func (ds *DataStore) AddTracklist(tx *sql.Tx, tracklist *Tracklist) error {
	_, err := tx.Exec(sqlAddTracklist,
		tracklist.ID,
		tracklist.Name,
		tracklist.Date,
		tracklist.Created,
		tracklist.Updated)

	return errors.Wrap(err, "tx exec failed")
}

// RemoveTracklist removes a tracklist from the database.
func (ds *DataStore) RemoveTracklist(tx *sql.Tx, id string) error {
	_, err := tx.Exec(sqlRemoveTracklist, id)

	return errors.Wrap(err, "tx exec failed")
}

// GetTracklist selects a tracklist from the database with the given ID.
func (ds *DataStore) GetTracklist(id string) (*Tracklist, error) {
	var tracklist Tracklist

	err := ds.QueryRowx(sqlGetTracklistByID, id).StructScan(&tracklist)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, errors.Wrap(err, "db query row failed")
	default:
		return &tracklist, nil
	}
}

// GetTracklistWithTracks selects a tracklist, and the tracks in the tracklist
// from the database with the given ID.
func (ds *DataStore) GetTracklistWithTracks(id string) (*Tracklist, error) {
	rows, err := ds.Queryx(sqlGetTracklistWithTracksByID, id)
	if err != nil {
		return nil, errors.Wrap(err, "db query failed")
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
			&track.ID,
			&track.Artist,
			&track.Name,
			&track.Genre,
			&track.BPM,
			&track.Key,
			&track.Created,
			&track.Updated)

		if err != nil {
			return nil, errors.Wrap(err, "rows scan failed")
		}

		tracklist.Tracks = append(tracklist.Tracks, &track)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows next failed")
	}

	return &tracklist, nil
}

// FindTracklistByName finds a tracklist from the database with the given name.
func (ds *DataStore) FindTracklistByName(name string) (*Tracklist, error) {
	var tracklist Tracklist

	err := ds.QueryRowx(sqlFindTracklistByName, name).StructScan(&tracklist)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, errors.Wrap(err, "db query row failed")
	default:
		return &tracklist, nil
	}
}

// FindTracklistWithTracksByName find a tracklist, and the tracks from the
// database with the given name.
func (ds *DataStore) FindTracklistWithTracksByName(name string) (*Tracklist, error) {
	rows, err := ds.Queryx(sqlFindTracklistWithTracksByName, name)
	if err != nil {
		return nil, errors.Wrap(err, "db query failed")
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
			&track.ID,
			&track.Artist,
			&track.Name,
			&track.Genre,
			&track.BPM,
			&track.Key,
			&track.Created,
			&track.Updated)

		if err != nil {
			return nil, errors.Wrap(err, "rows scan failed")
		}

		tracklist.Tracks = append(tracklist.Tracks, &track)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows next failed")
	}

	if tracklist.Name == "" {
		return nil, fmt.Errorf("could not find tracklist: %q", name)
	}

	return &tracklist, nil
}
