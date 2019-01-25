package services

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/tombell/memoir/database"
)

// TrackImport ...
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
func NewTrack(track *database.TrackRecord) *Track {
	return &Track{
		ID:      track.ID,
		Artist:  track.Artist,
		Name:    track.Name,
		Genre:   track.Genre,
		BPM:     track.BPM,
		Key:     track.Key,
		Created: track.Created,
		Updated: track.Updated,
	}
}

// ImportTrack ...
func (s *Services) ImportTrack(tx *sql.Tx, trackImport *TrackImport) (*Track, error) {
	track, err := s.DB.FindTrack(trackImport.Artist, trackImport.Name)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "find track failed")
	}

	if track == nil {
		id, _ := uuid.NewV4()
		bpm, _ := strconv.Atoi(trackImport.BPM)

		track = &database.TrackRecord{
			ID:      id.String(),
			Name:    trackImport.Name,
			Artist:  trackImport.Artist,
			BPM:     bpm,
			Key:     trackImport.Key,
			Genre:   trackImport.Genre,
			Created: time.Now().UTC(),
			Updated: time.Now().UTC(),
		}

		if err := s.DB.InsertTrack(tx, track); err != nil {
			tx.Rollback()
			return nil, errors.Wrap(err, "insert track failed")
		}
	}

	return NewTrack(track), nil
}
