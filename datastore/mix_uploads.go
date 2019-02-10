package datastore

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	sqlAddMixUpload = `
		INSERT INTO mix_uploads (
			id,
			tracklist_id,
			filename,
			location,
			created,
			updated
		) VALUES ($1, $2, $3, $4, $5, $6)`

	sqlRemoveMixUpload = `
		DELETE FROM mix_uploads
		WHERE id = $1`
)

// MixUpload represents a single mix upload row in the database.
type MixUpload struct {
	ID          string
	TracklistID string
	Filename    string
	Location    string
	Created     time.Time
	Updated     time.Time
}

// AddMixUpload ...
func (ds *DataStore) AddMixUpload(tx *sqlx.Tx, mix *MixUpload) error {
	_, err := tx.Exec(sqlAddMixUpload,
		mix.ID,
		mix.TracklistID,
		mix.Filename,
		mix.Location,
		mix.Created,
		mix.Updated)

	return errors.Wrap(err, "tx exec failed")
}

// RemoveMixUpload ...
func (ds *DataStore) RemoveMixUpload(tx *sqlx.Tx, id string) error {
	_, err := tx.Exec(sqlRemoveMixUpload, id)

	return errors.Wrap(err, "tx exec failed")
}
