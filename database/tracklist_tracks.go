package database

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// TracklistTrackRecord represents a single tracklist_track row in the database.
// Used for mapping a track to a tracklist.
type TracklistTrackRecord struct {
	ID          string `db:"id"`
	TracklistID string `db:"tracklist_id"`
	TrackID     string `db:"track_id"`
	TrackNumber int    `db:"track_number"`
}

// InsertTracklistToTrack inserts a new tracklist to track mapping into the
// database.
func (db *Database) InsertTracklistToTrack(tx *sqlx.Tx, tracklistTrack *TracklistTrackRecord) error {
	_, err := tx.Exec(sqlInsertTracklistTrack,
		tracklistTrack.ID,
		tracklistTrack.TracklistID,
		tracklistTrack.TrackID,
		tracklistTrack.TrackNumber)

	return errors.Wrap(err, "tx exec failed")
}
