package services

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/gofrs/uuid"

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
func NewTrackFromSearchResult(record *datastore.TrackSearchResult) *Track {
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

// ImportTrack imports the new track if it doesn't already exist.
func (s *Services) ImportTrack(tx *sql.Tx, trackImport *TrackImport) (*Track, error) {
	track, err := s.DataStore.FindTrackByArtistAndName(trackImport.Artist, trackImport.Name)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("find track failed: %w", err)
	}

	if track == nil {
		id, _ := uuid.NewV4()
		bpm, _ := strconv.ParseFloat(trackImport.BPM, 64)

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
			return nil, fmt.Errorf("insert track failed: %w", err)
		}
	}

	return NewTrack(track), nil
}

// GetMostPlayedTracks gets the tracks that have been played most in tracklists.
func (s *Services) GetMostPlayedTracks(limit int) ([]*Track, error) {
	tracks, err := s.DataStore.FindMostPlayedTracks(limit)
	if err != nil {
		return nil, fmt.Errorf("find most played tracks failed: %w", err)
	}

	var models []*Track

	for _, track := range tracks {
		models = append(models, NewTrack(track))
	}

	return models, nil
}

// SearchTracks searches for tracks that have artists and/or names matching the
// query.
func (s *Services) SearchTracks(query string) ([]*Track, error) {
	tracks, err := s.DataStore.FindTracksByQuery(query + ":*")
	if err != nil {
		return nil, fmt.Errorf("find tracks by query failed: %w", err)
	}

	var models []*Track

	for _, track := range tracks {
		models = append(models, NewTrackFromSearchResult(track))
	}

	return models, nil
}
