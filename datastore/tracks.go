package datastore

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	sqlAddTrack = `
		INSERT INTO tracks (
			id,
			artist,
			name,
			genre,
			bpm,
			key,
			created,
			updated
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	sqlGetTrackByID = `
		SELECT
			id,
			artist,
			name,
			genre,
			bpm,
			key
			created,
			updated
		FROM tracks
		WHERE id = $1
		LIMIT 1`

	sqlFindTrackByArtistAndName = `
		SELECT
			id,
			artist,
			name,
			genre,
			bpm,
			key,
			created,
			updated
		FROM tracks
		WHERE artist = $1
		AND name = $2
		LIMIT 1`
)

// Track ...
type Track struct {
	ID      string
	Artist  string
	Name    string
	Genre   string
	BPM     int
	Key     string
	Created time.Time
	Updated time.Time
}

// AddTrack ...
func (ds *DataStore) AddTrack(tx *sqlx.Tx, track *Track) error {
	_, err := tx.Exec(sqlAddTrack,
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

// GetTrack ...
func (ds *DataStore) GetTrack(id string) (*Track, error) {
	var track Track

	err := ds.QueryRowx(sqlGetTrackByID, id).StructScan(&track)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, errors.Wrap(err, "query row failed")
	default:
		return &track, nil
	}
}

// FindTrackByArtistAndName ...
func (ds *DataStore) FindTrackByArtistAndName(artist, name string) (*Track, error) {
	var track Track

	err := ds.QueryRowx(sqlFindTrackByArtistAndName, artist, name).StructScan(&track)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, errors.Wrap(err, "query row failed")
	default:
		return &track, nil
	}
}
