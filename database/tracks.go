package database

import "database/sql"

// TrackRecord  represents a single track row in the database.
type TrackRecord struct {
	ID     string
	Artist string
	Name   string
	Genre  string
	BPM    int
	Key    string
}

// InsertTrack inserts a new track into the database.
func (db *Database) InsertTrack(tx *sql.Tx, track *TrackRecord) error {
	_, err := tx.Exec(sqlInsertTrack,
		track.ID,
		track.Artist,
		track.Name,
		track.Genre,
		track.BPM,
		track.Key)

	return err
}

// GetTrack returns a single track with the given ID from the database.
// Returns nil if the track doesn't exist.
func (db *Database) GetTrack(id string) (*TrackRecord, error) {
	var track TrackRecord

	err := db.conn.QueryRow(sqlGetTrackByID, id).Scan(
		&track.ID,
		&track.Artist,
		&track.Name,
		&track.Genre,
		&track.BPM,
		&track.Key)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &track, nil
	}
}
