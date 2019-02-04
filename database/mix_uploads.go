package database

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// MixUploadRecord represents a single mix upload row in the database.
type MixUploadRecord struct {
	ID          string    `db:"id"`
	TracklistID string    `db:"tracklist_id"`
	Filename    string    `db:"filename"`
	Location    string    `db:"location"`
	Created     time.Time `db:"created"`
	Updated     time.Time `db:"updated"`
}

// InsertMixUpload inserts a new mix upload into the database.
func (db *Database) InsertMixUpload(tx *sqlx.Tx, mix *MixUploadRecord) error {
	_, err := tx.Exec(sqlInsertMixUpload,
		mix.ID,
		mix.TracklistID,
		mix.Filename,
		mix.Location,
		mix.Created,
		mix.Updated)

	return errors.Wrap(err, "tx exec failed")
}

// RemoveMixUpload removes a mix upload with the given ID from the database.
func (db *Database) RemoveMixUpload(tx *sqlx.Tx, id string) error {
	_, err := tx.Exec(sqlRemoveMixUpload, id)
	return errors.Wrap(err, "tx exec failed")
}
