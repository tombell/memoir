package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tombell/tonality"

	"github.com/tombell/memoir/internal/datastore"
)

func (s *Services) GetTracklists(page, limit int32) ([]*Tracklist, int64, error) {
	s.Logger.Info("get-tracklists:started", "page", page)
	defer s.Logger.Info("get-tracklists:finished", "page", page)

	var total int64

	total, err := s.DataStore.CountTracklists(context.Background())
	if err != nil {
		s.Logger.Error("get-tracklists:error", "msg", "datastore count tracklists failed", "err", err)
		return nil, -1, fmt.Errorf("count tracklists failed: %w", ErrQueryFailed)
	}

	rows, err := s.DataStore.GetTracklists(context.Background(), datastore.GetTracklistsParams{
		Offset: limit * (page - 1),
		Limit:  limit,
	})
	if err != nil {
		s.Logger.Error("get-tracklists:error", "msg", "datastore get tracklists failed", "err", err)
		return nil, -1, fmt.Errorf("get tracklists failed: %w", ErrQueryFailed)
	}

	tracklists := make([]*Tracklist, 0, len(rows))

	for _, row := range rows {
		tracklists = append(tracklists, &Tracklist{
			ID:         row.ID,
			Name:       row.Name,
			Artwork:    row.Artwork,
			URL:        row.URL,
			Date:       row.Date,
			Created:    row.Created,
			Updated:    row.Updated,
			TrackCount: int(row.TrackCount),
		})
	}

	return tracklists, total, nil
}

func (s *Services) GetTracklist(id string) (*Tracklist, error) {
	s.Logger.Info("get-tracklist:started", "id", id)
	defer s.Logger.Info("get-tracklist:finished", "id", id)

	rows, err := s.DataStore.GetTracklistWithTracks(context.Background(), id)
	if err != nil {
		s.Logger.Error("get-tracklists:error", "msg", "datastore get tracklist with tracks failed", "err", err)
		return nil, fmt.Errorf("get tracklists with tracks failed: %w", ErrQueryFailed)
	}

	if len(rows) == 0 {
		s.Logger.Info("get-tracklists", "msg", "tracklist not found")
		return nil, nil
	}

	tracklist := &Tracklist{
		ID:      rows[0].ID,
		Name:    rows[0].Name,
		Artwork: rows[0].Artwork,
		URL:     rows[0].URL,
		Date:    rows[0].Date,
		Created: rows[0].Created,
		Updated: rows[0].Updated,
	}

	for _, row := range rows {
		tracklist.Tracks = append(tracklist.Tracks, &Track{
			ID:      row.TrackID,
			Artist:  row.Artist,
			Name:    row.TrackName,
			Genre:   row.Genre,
			BPM:     row.BPM,
			Key:     row.Key,
			Created: row.TrackCreated,
			Updated: row.TrackUpdated,
		})
	}

	tracklist.TrackCount = len(tracklist.Tracks)

	return tracklist, nil
}

func (s *Services) AddTracklist(model *TracklistAdd) (*Tracklist, error) {
	s.Logger.Info("add-tracklist:started", "name", model.Name)
	defer s.Logger.Info("add-tracklist:finished", "name", model.Name)

	tx, err := s.DataStore.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("db begin failed: %w", err)
	}

	queries := s.DataStore.WithTx(tx)

	tracklist, err := queries.AddTracklist(context.Background(), datastore.AddTracklistParams{
		ID:      uuid.NewString(),
		Name:    model.Name,
		Date:    model.Date,
		URL:     model.URL,
		Artwork: model.Artwork,
		Created: time.Now().UTC(),
		Updated: time.Now().UTC(),
	})
	if err != nil {
		tx.Rollback(context.Background())

		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				s.Logger.Error("add-tracklist:error", "tracklist with name already exists", "name", model.Name)
				return nil, ErrTracklistExists
			}
		}

		return nil, ErrQueryFailed

	}

	for idx, data := range model.Tracks {
		normalizedKey, err := tonality.ConvertKeyToNotation(data[3], tonality.CamelotKeys)
		if err != nil {
			tx.Rollback(context.Background())
			return nil, fmt.Errorf("normalizing key to camelot key notation failed: %w", err)
		}

		row, err := queries.GetTrackByArtistAndName(context.Background(), datastore.GetTrackByArtistAndNameParams{
			Artist: data[1],
			Name:   data[0],
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get track by artist and name failed: %w", err)
		}

		foundTrackID := row.ID

		if errors.Is(err, pgx.ErrNoRows) {
			foundTrackID = uuid.NewString()
			bpm, _ := strconv.ParseFloat(data[2], 64)

			if err := queries.AddTrack(context.Background(), datastore.AddTrackParams{
				ID:      foundTrackID,
				Name:    data[0],
				Artist:  data[1],
				BPM:     bpm,
				Key:     normalizedKey,
				Genre:   data[4],
				Created: time.Now().UTC(),
				Updated: time.Now().UTC(),
			}); err != nil {
				tx.Rollback(context.Background())
				return nil, fmt.Errorf("add track failed: %w", err)
			}
		}

		if err := queries.AddTracklistTrack(context.Background(), datastore.AddTracklistTrackParams{
			ID:          uuid.NewString(),
			TracklistID: tracklist.ID,
			TrackID:     foundTrackID,
			TrackNumber: int32(idx + 1),
		}); err != nil {
			tx.Rollback(context.Background())
			return nil, fmt.Errorf("insert tracklist_track failed: %w", err)
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		tx.Rollback(context.Background())
		return nil, fmt.Errorf("tx commit failed: %w", err)
	}

	return &Tracklist{
		ID:      tracklist.ID,
		Name:    tracklist.Name,
		Date:    tracklist.Date,
		Artwork: tracklist.Artwork,
		URL:     tracklist.URL,
		Created: tracklist.Created,
		Updated: tracklist.Updated,
	}, nil
}

func (s *Services) UpdateTracklist(id string, model *TracklistUpdate) (*Tracklist, error) {
	s.Logger.Info("update-tracklist:started", "id", id)
	defer s.Logger.Info("update-tracklist:finished", "id", id)

	tx, err := s.DataStore.Begin(context.Background())
	if err != nil {
		// TODO: log error
		// return nil, fmt.Errorf("db begin failed: %w", err)
		return nil, ErrQueryFailed
	}

	queries := s.DataStore.WithTx(tx)

	if _, err = queries.UpdateTracklist(context.Background(), datastore.UpdateTracklistParams{
		ID:   id,
		Name: model.Name,
		Date: model.Date,
		URL:  model.URL,
	}); err != nil {
		tx.Rollback(context.Background())
		// TODO: log error
		// return nil, fmt.Errorf("update tracklist failed: %w", err)
		return nil, ErrQueryFailed
	}

	if err := tx.Commit(context.Background()); err != nil {
		tx.Rollback(context.Background())
		// TODO: log error
		// return nil, fmt.Errorf("tx commit failed: %w", err)
		return nil, ErrQueryFailed
	}

	rows, err := s.DataStore.GetTracklistWithTracks(context.Background(), id)
	if err != nil {
		// TODO: log error
		// return nil, fmt.Errorf("find tracklist failed: %w", err)
		return nil, ErrQueryFailed
	}

	tracklist := &Tracklist{
		ID:      rows[0].ID,
		Name:    rows[0].Name,
		Artwork: rows[0].Artwork,
		URL:     rows[0].URL,
		Date:    rows[0].Date,
		Created: rows[0].Created,
		Updated: rows[0].Updated,
	}

	for _, row := range rows {
		tracklist.Tracks = append(tracklist.Tracks, &Track{
			ID:      row.TrackID,
			Artist:  row.Artist,
			Name:    row.TrackName,
			Genre:   row.Genre,
			BPM:     row.BPM,
			Key:     row.Key,
			Created: row.TrackCreated,
			Updated: row.TrackUpdated,
		})
	}

	tracklist.TrackCount = len(tracklist.Tracks)

	return tracklist, nil
}

func (s *Services) GetTracklistsByTrack(id string, page, limit int32) ([]*Tracklist, int64, error) {
	s.Logger.Info("get-tracklists-by-track:started", "id", id, "page", page)
	defer s.Logger.Info("get-tracklists-by-track:finished", "id", id, "page", page)

	var total int64
	total, err := s.DataStore.CountTracklistsByTrack(context.Background(), id)
	if err != nil {
		return nil, -1, fmt.Errorf("datastore count tracklists by track failed: %w", err)
	}

	rows, err := s.DataStore.GetTracklistsByTrack(context.Background(), datastore.GetTracklistsByTrackParams{
		TrackID: id,
		Offset:  limit * (page - 1),
		Limit:   limit,
	})
	if err != nil {
		return nil, -1, fmt.Errorf("find tracklists by track id failed: %w", err)
	}

	tracklists := make([]*Tracklist, 0)

	for _, row := range rows {
		tracklists = append(tracklists, &Tracklist{
			ID:         row.ID,
			Name:       row.Name,
			Date:       row.Date,
			URL:        row.URL,
			Artwork:    row.Artwork,
			Created:    row.Created,
			Updated:    row.Updated,
			TrackCount: int(row.TrackCount),
		})
	}

	return tracklists, total, nil
}
