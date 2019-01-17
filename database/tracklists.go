package database

import (
	"database/sql"
	"time"
)

// TracklistRecord represents a single tracklist row in the database.
type TracklistRecord struct {
	ID      string
	Created time.Time
	Updated time.Time
}

// InsertTracklist inserts a new tracklist into the database.
func (db *Database) InsertTracklist(tx *sql.Tx, tracklist *TracklistRecord) error {
	_, err := tx.Exec(sqlInsertTracklist,
		tracklist.ID,
		tracklist.Created,
		tracklist.Updated)

	return err
}

// GetTracklist returns a single tracklist with the given ID from the database.
// Returns nil if the tracklist doesn't exist.
func (db *Database) GetTracklist(id string) (*TracklistRecord, error) {
	var tracklist TracklistRecord

	err := db.conn.QueryRow(sqlGetTracklistByID, id).Scan(
		&tracklist.ID,
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
