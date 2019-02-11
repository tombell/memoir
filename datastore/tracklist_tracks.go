package datastore

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
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

// TracklistTrack represents a single tracklist_track row in the database.
type TracklistTrack struct {
	ID          string
	TracklistID string
	TrackID     string
	TrackNumber int
}

// AddTracklistTrack adds a new tracklist to track mapping into the database.
func (ds *DataStore) AddTracklistTrack(tx *sqlx.Tx, tracklistTrack *TracklistTrack) error {
	_, err := tx.Exec(sqlAddTracklistTrack,
		tracklistTrack.ID,
		tracklistTrack.TracklistID,
		tracklistTrack.TrackID,
		tracklistTrack.TrackNumber)

	return errors.Wrap(err, "tx exec failed")
}
