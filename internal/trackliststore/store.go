package trackliststore

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/tombell/memoir/internal/datastore"
	db "github.com/tombell/memoir/internal/datastore/database"
	"github.com/tombell/memoir/internal/trackstore"
)

type Tracklist struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Date    time.Time `json:"date"`
	URL     string    `json:"url"`
	Artwork string    `json:"artwork"`

	Created time.Time `json:"-"`
	Updated time.Time `json:"-"`

	Tracks     []*trackstore.Track `json:"tracks,omitempty"`
	TrackCount int                 `json:"trackCount"`
}

type AddTracklistParams struct {
	Name    string     `json:"name"`
	Date    time.Time  `json:"date"`
	URL     string     `json:"url"`
	Artwork string     `json:"artwork"`
	Tracks  [][]string `json:"tracks"`
}

type UpdateTracklistParams struct {
	Name string    `json:"name"`
	Date time.Time `json:"date"`
	URL  string    `json:"url"`
}

type Store struct {
	dataStore *datastore.Store
}

func New(store *datastore.Store) *Store {
	return &Store{dataStore: store}
}

func (s *Store) GetTracklists(page, limit int32) ([]*Tracklist, int64, error) {
	var total int64

	total, err := s.dataStore.CountTracklists(context.Background())
	if err != nil {
		return nil, -1, fmt.Errorf("count tracklists failed: %w", err)
	}

	rows, err := s.dataStore.GetTracklists(context.Background(), db.GetTracklistsParams{
		Offset: limit * (page - 1),
		Limit:  limit,
	})
	if err != nil {
		return nil, -1, fmt.Errorf("get tracklists failed: %w", err)
	}

	tracklists := make([]*Tracklist, 0, len(rows))

	for _, row := range rows {
		tracklists = append(tracklists, &Tracklist{
			ID:         row.Tracklist.ID,
			Name:       row.Tracklist.Name,
			Artwork:    row.Tracklist.Artwork,
			URL:        row.Tracklist.URL,
			Date:       row.Tracklist.Date,
			Created:    row.Tracklist.Created,
			Updated:    row.Tracklist.Updated,
			TrackCount: int(row.TrackCount),
		})
	}

	return tracklists, total, nil
}

func (s *Store) GetTracklist(id string) (*Tracklist, error) {
	rows, err := s.dataStore.GetTracklistWithTracks(context.Background(), id)
	if err != nil {
		return nil, fmt.Errorf("get tracklists with tracks failed: %w", err)
	}

	if len(rows) == 0 {
		return nil, nil
	}

	tracklist := &Tracklist{
		ID:      rows[0].Tracklist.ID,
		Name:    rows[0].Tracklist.Name,
		Artwork: rows[0].Tracklist.Artwork,
		URL:     rows[0].Tracklist.URL,
		Date:    rows[0].Tracklist.Date,
		Created: rows[0].Tracklist.Created,
		Updated: rows[0].Tracklist.Updated,
	}

	for _, row := range rows {
		tracklist.Tracks = append(tracklist.Tracks, &trackstore.Track{
			ID:      row.Track.ID,
			Artist:  row.Track.Artist,
			Name:    row.Track.Name,
			Genre:   row.Track.Genre,
			BPM:     row.Track.BPM,
			Key:     row.Track.Key,
			Created: row.Track.Created,
			Updated: row.Track.Updated,
		})
	}

	tracklist.TrackCount = len(tracklist.Tracks)

	return tracklist, nil
}

func (s *Store) AddTracklist(model *AddTracklistParams) (*Tracklist, error) {
	tx, err := s.dataStore.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("db begin failed: %w", err)
	}
	defer tx.Rollback(context.Background())

	queries := s.dataStore.WithTx(tx)

	tracklist, err := queries.AddTracklist(context.Background(), db.AddTracklistParams{
		ID:      uuid.NewString(),
		Name:    model.Name,
		Date:    model.Date,
		URL:     model.URL,
		Artwork: model.Artwork,
	})
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return nil, nil
			}
		}

		return nil, err

	}

	for idx, data := range model.Tracks {
		row, err := queries.GetTrackByArtistAndName(context.Background(), db.GetTrackByArtistAndNameParams{
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

			if err := queries.AddTrack(context.Background(), db.AddTrackParams{
				ID:      foundTrackID,
				Name:    data[0],
				Artist:  data[1],
				BPM:     bpm,
				Key:     strings.ToUpper(data[3]),
				Genre:   data[4],
				Created: time.Now().UTC(),
				Updated: time.Now().UTC(),
			}); err != nil {
				return nil, fmt.Errorf("add track failed: %w", err)
			}
		}

		if err := queries.AddTracklistTrack(context.Background(), db.AddTracklistTrackParams{
			ID:          uuid.NewString(),
			TracklistID: tracklist.ID,
			TrackID:     foundTrackID,
			TrackNumber: int32(idx + 1),
		}); err != nil {
			return nil, fmt.Errorf("insert tracklist_track failed: %w", err)
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
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

func (s *Store) UpdateTracklist(id string, model *UpdateTracklistParams) (*Tracklist, error) {
	tx, err := s.dataStore.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("db begin failed: %w", err)
	}
	defer tx.Rollback(context.Background())

	queries := s.dataStore.WithTx(tx)

	if _, err = queries.UpdateTracklist(context.Background(), db.UpdateTracklistParams{
		ID:   id,
		Name: model.Name,
		Date: model.Date,
		URL:  model.URL,
	}); err != nil {
		return nil, fmt.Errorf("update tracklist failed: %w", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return nil, fmt.Errorf("tx commit failed: %w", err)
	}

	rows, err := s.dataStore.GetTracklistWithTracks(context.Background(), id)
	if err != nil {
		return nil, fmt.Errorf("find tracklist failed: %w", err)
	}

	tracklist := &Tracklist{
		ID:      rows[0].Tracklist.ID,
		Name:    rows[0].Tracklist.Name,
		Artwork: rows[0].Tracklist.Artwork,
		URL:     rows[0].Tracklist.URL,
		Date:    rows[0].Tracklist.Date,
		Created: rows[0].Tracklist.Created,
		Updated: rows[0].Tracklist.Updated,
	}

	for _, row := range rows {
		tracklist.Tracks = append(tracklist.Tracks, &trackstore.Track{
			ID:      row.Track.ID,
			Artist:  row.Track.Artist,
			Name:    row.Track.Name,
			Genre:   row.Track.Genre,
			BPM:     row.Track.BPM,
			Key:     row.Track.Key,
			Created: row.Track.Created,
			Updated: row.Track.Updated,
		})
	}

	tracklist.TrackCount = len(tracklist.Tracks)

	return tracklist, nil
}

func (s *Store) GetTracklistsByTrack(id string, page, limit int32) ([]*Tracklist, int64, error) {
	var total int64
	total, err := s.dataStore.CountTracklistsByTrack(context.Background(), id)
	if err != nil {
		return nil, -1, fmt.Errorf("datastore count tracklists by track failed: %w", err)
	}

	rows, err := s.dataStore.GetTracklistsByTrack(context.Background(), db.GetTracklistsByTrackParams{
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
			ID:         row.Tracklist.ID,
			Name:       row.Tracklist.Name,
			Date:       row.Tracklist.Date,
			URL:        row.Tracklist.URL,
			Artwork:    row.Tracklist.Artwork,
			Created:    row.Tracklist.Created,
			Updated:    row.Tracklist.Updated,
			TrackCount: int(row.TrackCount),
		})
	}

	return tracklists, total, nil
}
