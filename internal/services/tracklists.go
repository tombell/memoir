package services

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/tombell/tonality"

	"github.com/tombell/memoir/internal/datastore"
	"github.com/tombell/memoir/internal/services/models"
)

func (s *Services) GetTracklists(page, limit int) ([]*models.Tracklist, int, error) {
	s.Logger.Info("get-tracklists", "page", page)

	done := make(chan struct{})

	var count int

	go func() {
		count, _ = s.DataStore.GetTracklistsCount()
		close(done)
	}()

	tracklists, err := s.DataStore.GetTracklists(limit*(page-1), limit)
	if err != nil {
		return nil, -1, fmt.Errorf("get tracklists failed: %w", err)
	}

	m := make([]*models.Tracklist, 0)

	for _, tracklist := range tracklists {
		m = append(m, models.NewTracklist(tracklist))
	}

	<-done

	return m, count, nil
}

func (s *Services) AddTracklist(model *models.TracklistAdd) (*models.Tracklist, error) {
	s.Logger.Info("add-tracklist", "name", model.Name)

	tracklist, err := s.DataStore.FindTracklistByName(model.Name)
	if err != nil {
		return nil, fmt.Errorf("find tracklist failed: %w", err)
	}
	if tracklist != nil {
		return nil, fmt.Errorf("tracklist named %q already exists", model.Name)
	}

	id, _ := uuid.NewV4()

	tracklist = &datastore.Tracklist{
		ID:      id.String(),
		Name:    model.Name,
		Date:    model.Date,
		URL:     model.URL,
		Artwork: model.Artwork,
		Created: time.Now().UTC(),
		Updated: time.Now().UTC(),
	}

	tx, err := s.DataStore.Begin()
	if err != nil {
		return nil, fmt.Errorf("db begin failed: %w", err)
	}

	if err := s.DataStore.AddTracklist(tx, tracklist); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("insert tracklist failed: %w", err)
	}

	for idx, data := range model.Tracks {
		normalizedKey, err := tonality.ConvertKeyToNotation(data[3], tonality.CamelotKeys)
		if err != nil {
			return nil, fmt.Errorf("normalizing key to camelot key notation failed: %w", err)
		}

		trackImport := &models.TrackAdd{
			Name:   data[0],
			Artist: data[1],
			BPM:    data[2],
			Key:    normalizedKey,
			Genre:  data[4],
		}

		track, err := s.AddTrack(tx, trackImport)
		if err != nil {
			return nil, fmt.Errorf("add track failed: %w", err)
		}

		id, _ := uuid.NewV4()

		tracklistTrack := &datastore.TracklistTrack{
			ID:          id.String(),
			TracklistID: tracklist.ID,
			TrackID:     track.ID,
			TrackNumber: idx + 1,
		}

		if err := s.DataStore.AddTracklistTrack(tx, tracklistTrack); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("insert tracklist_track failed: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("tx commit failed: %w", err)
	}

	return models.NewTracklist(tracklist), nil
}

func (s *Services) UpdateTracklist(id string, model *models.TracklistUpdate) (*models.Tracklist, error) {
	s.Logger.Info("update-tracklist", "id", id)

	tx, err := s.DataStore.Begin()
	if err != nil {
		return nil, fmt.Errorf("db begin failed: %w", err)
	}

	if err := s.DataStore.UpdateTracklist(tx, id, model.Name, model.URL, model.Date); err != nil {
		return nil, fmt.Errorf("update tracklist failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("tx commit failed: %w", err)
	}

	tracklist, err := s.DataStore.FindTracklistWithTracks(id)
	if err != nil {
		return nil, fmt.Errorf("find tracklist failed: %w", err)
	}
	if tracklist == nil {
		return nil, fmt.Errorf("tracklist %q exist", id)
	}

	return models.NewTracklist(tracklist), nil
}

func (s *Services) GetTracklist(id string) (*models.Tracklist, error) {
	s.Logger.Info("get-tracklist", "id", id)

	tracklist, err := s.DataStore.FindTracklistWithTracks(id)
	if err != nil {
		return nil, fmt.Errorf("get tracklist with tracks failed: %w", err)
	}
	if tracklist == nil {
		return nil, nil
	}

	return models.NewTracklist(tracklist), nil
}

func (s *Services) GetTracklistsByTrack(id string, page, limit int) ([]*models.Tracklist, int, error) {
	s.Logger.Info("get-tracklists-by-track", "id", id, "page", page)

	done := make(chan struct{})

	var count int

	go func() {
		count, _ = s.DataStore.FindTracklistsByTrackIDCount(id)
		close(done)
	}()

	tracklists, err := s.DataStore.FindTracklistsByTrackID(id, limit*(page-1), limit)
	if err != nil {
		return nil, -1, fmt.Errorf("find tracklists by track id failed: %w", err)
	}

	m := make([]*models.Tracklist, 0)

	for _, tracklist := range tracklists {
		m = append(m, models.NewTracklist(tracklist))
	}

	<-done

	return m, count, nil
}