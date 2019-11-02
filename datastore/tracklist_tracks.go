package datastore

import (
	"database/sql"
	"fmt"
)

const (
	sqlAddTracklistTrack = `
		INSERT INTO tracklist_tracks (
			id,
			tracklist_id,
			track_id,
			track_number
		) VALUES ($1, $2, $3, $4)`
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
	_, err := tx.Exec(sqlAddTracklistTrack,
		tt.ID,
		tt.TracklistID,
		tt.TrackID,
		tt.TrackNumber)

	if err != nil {
		return fmt.Errorf("tx exec failed: %w", err)
	}

	return nil
}
