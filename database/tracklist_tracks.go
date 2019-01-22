package database

import "database/sql"

// TracklistTrackRecord represents a single tracklist_track row in the database.
// Used for mapping a track to a tracklist.
type TracklistTrackRecord struct {
	ID          string
	TracklistID string
	TrackID     string
	TrackNumber int
}

// InsertTracklistToTrack inserts a new tracklist to track mapping into the
// database.
func (db *Database) InsertTracklistToTrack(tx *sql.Tx, tracklistTrack *TracklistTrackRecord) error {
	_, err := tx.Exec(sqlInsertTracklistTrack,
		tracklistTrack.ID,
		tracklistTrack.TracklistID,
		tracklistTrack.TrackID,
		tracklistTrack.TrackNumber)

	return err
}
