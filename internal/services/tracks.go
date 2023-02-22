package services

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/gofrs/uuid"

	"github.com/tombell/memoir/internal/datastore"
	"github.com/tombell/memoir/internal/services/models"
)

func (s *Services) GetTrack(id string) (*models.Track, error) {
	s.Logger.Info("get-track", "id", id)

	track, err := s.DataStore.FindTrack(id)
	if err != nil {
		return nil, fmt.Errorf("get track failed: %w", err)
	}
	if track == nil {
		return nil, nil
	}

	return models.NewTrack(track), nil
}

func (s *Services) AddTrack(tx *sql.Tx, model *models.TrackAdd) (*models.Track, error) {
	s.Logger.Info("add-track", "name", model.Name)

	track, err := s.DataStore.FindTrackByArtistAndName(model.Artist, model.Name)
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

	return models.NewTrack(track), nil
}

func (s *Services) GetMostPlayedTracks(limit int) ([]*models.Track, error) {
	s.Logger.Info("get-most-played-tracks", "limit", limit)

	tracks, err := s.DataStore.FindMostPlayedTracks(limit)
	if err != nil {
		return nil, fmt.Errorf("find most played tracks failed: %w", err)
	}

	m := make([]*models.Track, 0)

	for _, track := range tracks {
		m = append(m, models.NewTrack(track))
	}

	return m, nil
}

func (s *Services) SearchTracks(query string, limit int) ([]*models.Track, error) {
	s.Logger.Info("search-tracks", "query", query)

	tracks, err := s.DataStore.FindTracksByQuery(query, limit)
	if err != nil {
		return nil, fmt.Errorf("find tracks by query failed: %w", err)
	}

	m := make([]*models.Track, 0)

	for _, track := range tracks {
		m = append(m, models.NewTrackFromSearchResult(track))
	}

	return m, nil
}
