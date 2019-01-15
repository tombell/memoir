package database

import (
	"database/sql"
	"time"
)

// TrackRecord  represents a single track row in the database.
type TrackRecord struct {
	ID      string
	Artist  string
	Name    string
	Genre   string
	BPM     int
	Key     string
	Created time.Time
	Updated time.Time
}

// InsertTrack inserts a new track into the database.
func (db *Database) InsertTrack(tx *sql.Tx, track *TrackRecord) error {
	_, err := tx.Exec(sqlInsertTrack,
		track.ID,
		track.Artist,
		track.Name,
		track.Genre,
		track.BPM,
		track.Key,
		track.Created,
		track.Updated)

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
		&track.Key,
		&track.Created,
		&track.Updated)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, err
	default:
		return &track, nil
	}
}

// GetTracksByGenre ...
func (db *Database) GetTracksByGenre(genre string) ([]*TrackRecord, error) {
	rows, err := db.conn.Query(sqlGetTracksByGenre, genre)
	if err != nil {
		return nil, err
	}

	var tracks []*TrackRecord

	for rows.Next() {
		var track TrackRecord

		err := rows.Scan(
			&track.ID,
			&track.Artist,
			&track.Name,
			&track.Genre,
			&track.BPM,
			&track.Key,
			&track.Created,
			&track.Updated)

		if err != nil {
			return nil, err
		}

		tracks = append(tracks, &track)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tracks, nil
}
