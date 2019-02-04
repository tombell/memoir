package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// TracklistRecord represents a single tracklist row in the database.
type TracklistRecord struct {
	ID      string    `db:"id"`
	Name    string    `db:"name"`
	Date    time.Time `db:"date"`
	Created time.Time `db:"created"`
	Updated time.Time `db:"updated"`

	Tracks []*TrackRecord
}

// InsertTracklist inserts a new tracklist into the database.
func (db *Database) InsertTracklist(tx *sqlx.Tx, tracklist *TracklistRecord) error {
	_, err := tx.Exec(sqlInsertTracklist,
		tracklist.ID,
		tracklist.Name,
		tracklist.Date,
		tracklist.Created,
		tracklist.Updated)

	return errors.Wrap(err, "tx exec failed")
}

// RemoveTracklist removes a tracklist with the given ID from the database.
func (db *Database) RemoveTracklist(tx *sqlx.Tx, id string) error {
	_, err := tx.Exec(sqlRemoveTracklist, id)
	return errors.Wrap(err, "tx exec failed")
}

// GetTracklist returns a single tracklist with the given ID from the database.
// Returns nil if the tracklist doesn't exist.
func (db *Database) GetTracklist(id string) (*TracklistRecord, error) {
	var tracklist TracklistRecord

	err := db.conn.QueryRowx(sqlGetTracklistByID, id).StructScan(&tracklist)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, errors.Wrap(err, "db query row failed")
	default:
		return &tracklist, nil
	}
}

// GetTracklistWithTracks returns a single tracklist with the given ID from the
// database. Populates the tracklist with all the tracks for the tracklist.
// Returns nil if the tracklist doesn't exist.
func (db *Database) GetTracklistWithTracks(id string) (*TracklistRecord, error) {
	rows, err := db.conn.Queryx(sqlGetTracklistWithTracksByID, id)
	if err != nil {
		return nil, errors.Wrap(err, "db query failed")
	}
	defer rows.Close()

	var tracklist TracklistRecord

	for rows.Next() {
		var track TrackRecord

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

// FindTracklist finds a tracklist with the given name in the database.
// Returns nil if no matching tracklist is found.
func (db *Database) FindTracklist(name string) (*TracklistRecord, error) {
	var tracklist TracklistRecord

	err := db.conn.QueryRowx(sqlGetTracklistByName, name).StructScan(&tracklist)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, errors.Wrap(err, "db query row failed")
	default:
		return &tracklist, nil
	}
}

// FindTracklistWithTracks finds a tracklist with the given name in the database.
// Populates the tracklist with all the tracks for the tracklist. Returns nil if
// no matching tracklist is found.
func (db *Database) FindTracklistWithTracks(name string) (*TracklistRecord, error) {
	rows, err := db.conn.Queryx(sqlGetTracklistWithTracksByName, name)
	if err != nil {
		return nil, errors.Wrap(err, "db query failed")
	}
	defer rows.Close()

	var tracklist TracklistRecord

	for rows.Next() {
		var track TrackRecord

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
