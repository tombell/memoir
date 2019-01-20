package database

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
)

// TracklistRecord represents a single tracklist row in the database.
type TracklistRecord struct {
	ID      string
	Name    string
	Date    time.Time
	Created time.Time
	Updated time.Time
}

// InsertTracklist inserts a new tracklist into the database.
func (db *Database) InsertTracklist(tx *sql.Tx, tracklist *TracklistRecord) error {
	_, err := tx.Exec(sqlInsertTracklist,
		tracklist.ID,
		tracklist.Name,
		tracklist.Date,
		tracklist.Created,
		tracklist.Updated)

	return err
}

// InsertTracklistToTrack inserts a new tracklist to track mapping into the
// database.
func (db *Database) InsertTracklistToTrack(tx *sql.Tx, tracklistID string, trackID string) error {
	id, _ := uuid.NewV4()

	_, err := tx.Exec(sqlInsertTracklistTrack,
		id.String(),
		tracklistID,
		trackID)

	return err
}

// GetTracklist returns a single tracklist with the given ID from the database.
// Returns nil if the tracklist doesn't exist.
func (db *Database) GetTracklist(id string) (*TracklistRecord, error) {
	var tracklist TracklistRecord

	err := db.conn.QueryRow(sqlGetTracklistByID, id).Scan(
		&tracklist.ID,
		&tracklist.Name,
		&tracklist.Date,
		&tracklist.Created,
		&tracklist.Updated)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &tracklist, nil
	}
}

// FindTracklist finds a tracklist with the given name in the database.
// Returns nil if no matching tracklist is found.
func (db *Database) FindTracklist(name string) (*TracklistRecord, error) {
	var tracklist TracklistRecord

	err := db.conn.QueryRow(sqlGetTracklistByName, name).Scan(
		&tracklist.ID,
		&tracklist.Name,
		&tracklist.Date,
		&tracklist.Created,
		&tracklist.Updated)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &tracklist, nil
	}
}
