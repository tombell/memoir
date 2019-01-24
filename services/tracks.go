package services

import (
	"time"

	"github.com/tombell/memoir/database"
)

// TrackImport ...
type TrackImport struct {
	Artist string
	Name   string
	BPM    string
	Genre  string
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
func (s *Services) ImportTrack(trackImport *TrackImport) (*TrackRecord, error) {
	return nil, nil
}
