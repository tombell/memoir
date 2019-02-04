package database

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// TrackRecord represents a single track row in the database.
type TrackRecord struct {
	ID      string    `db:"id"`
	Artist  string    `db:"artist"`
	Name    string    `db:"name"`
	Genre   string    `db:"genre"`
	BPM     int       `db:"bpm"`
	Key     string    `db:"key"`
	Created time.Time `db:"created"`
	Updated time.Time `db:"updated"`
}

// InsertTrack inserts a new track into the database.
func (db *Database) InsertTrack(tx *sqlx.Tx, track *TrackRecord) error {
	_, err := tx.Exec(sqlInsertTrack,
		track.ID,
		track.Artist,
		track.Name,
		track.Genre,
		track.BPM,
		track.Key,
		track.Created,
		track.Updated)

	return errors.Wrap(err, "tx exec failed")
}

// GetTrack returns a single track with the given ID from the database.
// Returns nil if the track doesn't exist.
func (db *Database) GetTrack(id string) (*TrackRecord, error) {
	var track TrackRecord

	err := db.conn.QueryRowx(sqlGetTrackByID, id).StructScan(&track)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, errors.Wrap(err, "query row failed")
	default:
		return &track, nil
	}
}

// FindTrack finds a track with the given artist and name in the database.
// Returns nil if no matching track is found.
func (db *Database) FindTrack(artist, name string) (*TrackRecord, error) {
	var track TrackRecord

	err := db.conn.QueryRowx(sqlGetTrackByArtistAndName, artist, name).StructScan(&track)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, errors.Wrap(err, "query row failed")
	default:
		return &track, nil
	}
}
