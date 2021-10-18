package services

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/gofrs/uuid"

	"github.com/tombell/memoir/pkg/datastore"
)

// GetTrack gets a track with the given ID.
func (s *Services) GetTrack(rid, id string) (*Track, error) {
	s.Logger.Printf("[%s] getting track (id %s)", rid, id)

	track, err := s.DataStore.GetTrack(id)
	if err != nil {
		return nil, fmt.Errorf("get track failed: %w", err)
	}
	if track == nil {
		return nil, nil
	}

	return NewTrack(track), nil
}

// AddTrack adds the new track if it doesn't already exist.
func (s *Services) AddTrack(rid string, tx *sql.Tx, model *TrackAdd) (*Track, error) {
	s.Logger.Printf("[%s] adding track (name %s)", rid, model.Name)

	track, err := s.DataStore.GetTrackByArtistAndName(model.Artist, model.Name)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("find track failed: %w", err)
	}

	if track == nil {
		id, _ := uuid.NewV4()
		bpm, _ := strconv.ParseFloat(model.BPM, 64)

		track = &datastore.Track{
			ID:      id.String(),
			Name:    model.Name,
			Artist:  model.Artist,
			BPM:     bpm,
			Key:     model.Key,
			Genre:   model.Genre,
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
func (s *Services) GetMostPlayedTracks(rid string, limit int) ([]*Track, error) {
	s.Logger.Printf("[%s] getting most played tracks (limit %d)", rid, limit)

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
func (s *Services) SearchTracks(rid, query string, limit int) ([]*Track, error) {
	s.Logger.Printf("[%s] searching tracks (query %q)", rid, query)

	tracks, err := s.DataStore.FindTracksByQuery(query, limit)
	if err != nil {
		return nil, fmt.Errorf("find tracks by query failed: %w", err)
	}

	var models []*Track

	for _, track := range tracks {
		models = append(models, NewTrackFromSearchResult(track))
	}

	return models, nil
}
