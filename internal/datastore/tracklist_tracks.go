package datastore

import (
	"database/sql"
	"fmt"

	"github.com/tombell/memoir/internal/datastore/queries"
)

// TracklistTrack contains data about a tracklist_track row in the database.
type TracklistTrack struct {
	ID          string
	TracklistID string
	TrackID     string
	TrackNumber int
}

// AddTracklistTrack adds a new tracklist to track mapping into the database.
func (s *Store) AddTracklistTrack(tx *sql.Tx, tt *TracklistTrack) error {
	_, err := tx.Exec(
		queries.AddTracklistTrack,
		tt.ID,
		tt.TracklistID,
		tt.TrackID,
		tt.TrackNumber,
	)

	if err != nil {
		return fmt.Errorf("tx exec failed: %w", err)
	}

	return nil
}
