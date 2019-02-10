package services

import (
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/tombell/memoir/datastore"
)

// TrackImport contains data about a track to import from a Serato CSV export.
type TrackImport struct {
	Artist string
	Name   string
	BPM    string
	Genre  string
	Key    string
}

// Track contains data about a specific track.
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

// NewTrack returns a nrw track with fields mapped from a database record.
func NewTrack(record *datastore.Track) *Track {
	return &Track{
		ID:      record.ID,
		Artist:  record.Artist,
		Name:    record.Name,
		Genre:   record.Genre,
		BPM:     record.BPM,
		Key:     record.Key,
		Created: record.Created,
		Updated: record.Updated,
	}
}

// ImportTrack imports the new track if it doesn't already exist in the
// database.
func (s *Services) ImportTrack(tx *sqlx.Tx, trackImport *TrackImport) (*Track, error) {
	track, err := s.DataStore.FindTrackByArtistAndName(trackImport.Artist, trackImport.Name)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "find track failed")
	}

	if track == nil {
		id, _ := uuid.NewV4()
		bpm, _ := strconv.Atoi(trackImport.BPM)

		track = &datastore.Track{
			ID:      id.String(),
			Name:    trackImport.Name,
			Artist:  trackImport.Artist,
			BPM:     bpm,
			Key:     trackImport.Key,
			Genre:   trackImport.Genre,
			Created: time.Now().UTC(),
			Updated: time.Now().UTC(),
		}

		if err := s.DataStore.AddTrack(tx, track); err != nil {
			tx.Rollback()
			return nil, errors.Wrap(err, "insert track failed")
		}
	}

	return NewTrack(track), nil
}
