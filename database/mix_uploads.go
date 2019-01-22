package database

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// MixUploadRecord represents a single mix upload row in the database.
type MixUploadRecord struct {
	ID          string
	TracklistID string
	Filename    string
	Location    string
	Created     time.Time
	Updated     time.Time
}

// InsertMixUpload inserts a new mix upload into the database.
func (db *Database) InsertMixUpload(tx *sql.Tx, mix *MixUploadRecord) error {
	_, err := tx.Exec(sqlInsertMixUpload,
		mix.ID,
		mix.TracklistID,
		mix.Filename,
		mix.Location,
		mix.Created,
		mix.Updated)

	return errors.Wrap(err, "insert mix_upload failed")
}
