package datastore

import (
	"database/sql"
	"fmt"
	"time"
)

// MixUpload contains data about a mix upload row in the database.
type MixUpload struct {
	ID          string
	TracklistID string
	Filename    string
	Location    string
	Created     time.Time
	Updated     time.Time
}

// AddMixUpload adds a new mix upload into the database.
func (s *Store) AddMixUpload(tx *sql.Tx, mix *MixUpload) error {
	_, err := tx.Exec(insertMixUploadSQL,
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
func (s *Store) RemoveMixUpload(tx *sql.Tx, id string) error {
	if _, err := tx.Exec(deleteMixUploadSQL, id); err != nil {
		return fmt.Errorf("tx exec failed: %w", err)
	}

	return nil
}