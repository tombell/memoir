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

// TracklistTrack represents a single tracklist_track row in the database. Used
// for mapping a track to a tracklist.
type TracklistTrack struct {
	ID          string
	TracklistID string
	TrackID     string
	TrackNumber int
}

// AddTracklistTrack ...
func (ds *DataStore) AddTracklistTrack(tx *sqlx.Tx, tracklistTrack *TracklistTrack) error {
	_, err := tx.Exec(sqlAddTracklistTrack,
		tracklistTrack.ID,
		tracklistTrack.TracklistID,
		tracklistTrack.TrackID,
		tracklistTrack.TrackNumber)

	return errors.Wrap(err, "tx exec failed")
}
