package database

import (
	"database/sql"
	"time"
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
func (db *Database) InsertMixUpload(tx *sql.Tx, upload *MixUploadRecord) error {
	_, err := tx.Exec(sqlInsertMixUpload,
		upload.ID,
		upload.TracklistID,
		upload.Filename,
		upload.Location,
		upload.Created,
		upload.Updated)

	return err
}
