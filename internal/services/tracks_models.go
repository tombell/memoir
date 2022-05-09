package services

import (
	"time"

	"github.com/tombell/memoir/internal/datastore"
)

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

type TrackAdd struct {
	Artist string
	Name   string
	BPM    string
	Genre  string
	Key    string
}

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

func NewTrackFromSearchResult(record *datastore.Track) *Track {
	return &Track{
		ID:                record.ID,
		Artist:            record.Artist,
		ArtistHighlighted: record.ArtistHighlighted,
		Name:              record.Name,
		NameHighlighted:   record.NameHighlighted,
		Genre:             record.Genre,
		BPM:               record.BPM,
		Key:               record.Key,
		Created:           record.Created,
		Updated:           record.Updated,
	}
}
