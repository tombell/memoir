package datastore

import (
	"fmt"

	"github.com/tombell/memoir/internal/datastore/queries"
)

type TracklistTrack struct {
	ID          string
	TracklistID string
	TrackID     string
	TrackNumber int
}

func (s *Store) AddTracklistTrack(tx *Tx, tt *TracklistTrack) error {
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
