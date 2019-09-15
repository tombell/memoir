package datastore

import (
	"database/sql"
	"fmt"
	"time"
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

// AddMixUpload adds a new mix upload into the database.
func (ds *DataStore) AddMixUpload(tx *sql.Tx, mix *MixUpload) error {
	_, err := tx.Exec(sqlAddMixUpload,
		mix.ID,
		mix.TracklistID,
		mix.Filename,
		mix.Location,
		mix.Created,
		mix.Updated)

	if err != nil {
		return fmt.Errorf("tx exec failed: %w", err)
	}

	return nil
}

// RemoveMixUpload removes a mix upload from the database.
func (ds *DataStore) RemoveMixUpload(tx *sql.Tx, id string) error {
	if _, err := tx.Exec(sqlRemoveMixUpload, id); err != nil {
		return fmt.Errorf("tx exec failed: %w", err)
	}

	return nil
}
