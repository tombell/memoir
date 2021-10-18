package services

import (
	"time"

	"github.com/tombell/memoir/pkg/datastore"
)

// TrackImport contains data about a track to import.
type TrackImport struct {
	Artist string
	Name   string
	BPM    string
	Genre  string
	Key    string
}

// Track contains data about a track, with optional played count and search
// result highlighting.
type Track struct {
	ID     string  `json:"id"`
	Artist string  `json:"artist"`
	Name   string  `json:"name"`
	Genre  string  `json:"genre"`
	BPM    float64 `json:"bpm"`
	Key    string  `json:"key"`

	Created time.Time `json:"-"`
	Updated time.Time `json:"-"`

	Played int `json:"played,omitempty"`

	ArtistHighlighted string `json:"artistHighlighted,omitempty"`
	NameHighlighted   string `json:"nameHighlighted,omitempty"`
}

// NewTrack returns a new Track with fields mapped from a track database record.
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
		Played:  record.Played,
	}
}

// NewTrackFromSearchResult returns a new Track with fields mapped from a track
// search result database record.
func NewTrackFromSearchResult(record *datastore.Track) *Track {
	return &Track{
		ID:                record.ID,
		Artist:            record.Artist,
		Name:              record.Name,
		Genre:             record.Genre,
		BPM:               record.BPM,
		Key:               record.Key,
		Created:           record.Created,
		Updated:           record.Updated,
		ArtistHighlighted: record.ArtistHighlighted,
		NameHighlighted:   record.NameHighlighted,
	}
}
