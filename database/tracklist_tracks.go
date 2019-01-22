package database

import (
	"database/sql"

	"github.com/pkg/errors"
)

// TracklistTrackRecord represents a single tracklist_track row in the database.
// Used for mapping a track to a tracklist.
type TracklistTrackRecord struct {
	ID          string
	TracklistID string
	TrackID     string
	TrackNumber int
}

// InsertTracklistToTrack inserts a new tracklist to track mapping into the
// database.
func (db *Database) InsertTracklistToTrack(tx *sql.Tx, tracklistTrack *TracklistTrackRecord) error {
	_, err := tx.Exec(sqlInsertTracklistTrack,
		tracklistTrack.ID,
		tracklistTrack.TracklistID,
		tracklistTrack.TrackID,
		tracklistTrack.TrackNumber)

	return errors.Wrap(err, "tx exec failed")
}
