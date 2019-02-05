package database

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

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
