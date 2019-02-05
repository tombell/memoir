package database

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

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
