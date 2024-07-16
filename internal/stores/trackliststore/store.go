package trackliststore

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	db "github.com/tombell/memoir/internal/database"
	"github.com/tombell/memoir/internal/errors"
	"github.com/tombell/memoir/internal/stores/datastore"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

type Store struct {
	dataStore *datastore.Store
}

func New(store *datastore.Store) *Store {
	return &Store{dataStore: store}
}

func (s *Store) GetTracklists(ctx context.Context, page, limit int32) ([]*Tracklist, int64, error) {
	op := errors.Op("trackliststore[get-tracklists]")

	var total int64

	total, err := s.dataStore.CountTracklists(ctx)
	if err != nil {
		return nil, -1, errors.E(op, errors.Strf("count tracklists failed: %w", err))
	}

	rows, err := s.dataStore.GetTracklists(ctx, db.GetTracklistsParams{
		Offset: limit * (page - 1),
		Limit:  limit,
	})
	if err != nil {
		return nil, -1, errors.E(op, errors.Strf("get tracklists failed: %w", err))
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

func (s *Store) GetTracklist(ctx context.Context, id string) (*Tracklist, error) {
	op := errors.Op("trackliststore[get-tracklist]")

	_, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.E(op, http.StatusNotFound)
	}

	rows, err := s.dataStore.GetTracklistWithTracks(ctx, id)
	if err != nil {
		return nil, errors.E(op, errors.Strf("get tracklists with tracks failed: %w", err))
	}

	if len(rows) == 0 {
		return nil, errors.E(op, http.StatusNotFound)
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

func (s *Store) AddTracklist(ctx context.Context, model *AddTracklistParams) (*Tracklist, error) {
	op := errors.Op("trackliststore[add-tracklist]")

	tx, err := s.dataStore.Begin(ctx)
	if err != nil {
		return nil, errors.E(op, errors.Strf("db begin failed: %w", err))
	}
	defer tx.Rollback(ctx)

	queries := s.dataStore.WithTx(tx)

	tracklist, err := queries.AddTracklist(ctx, db.AddTracklistParams{
		ID:      uuid.NewString(),
		Name:    model.Name,
		Date:    model.Date,
		URL:     model.URL,
		Artwork: model.Artwork,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.E(op, http.StatusUnprocessableEntity)
		}

		return nil, errors.E(op, errors.Strf("add tracklist failed: %w", err))

	}

	for idx, data := range model.Tracks {
		row, err := queries.GetTrackByArtistAndName(ctx, db.GetTrackByArtistAndNameParams{
			Artist: data[1],
			Name:   data[0],
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.E(op, errors.Strf("get track by artist and name failed: %w", err))
		}

		foundTrackID := row.ID

		if errors.Is(err, pgx.ErrNoRows) {
			foundTrackID = uuid.NewString()
			bpm, _ := strconv.ParseFloat(data[2], 64)

			if err := queries.AddTrack(ctx, db.AddTrackParams{
				ID:      foundTrackID,
				Name:    data[0],
				Artist:  data[1],
				BPM:     bpm,
				Key:     strings.ToUpper(data[3]),
				Genre:   data[4],
				Created: time.Now().UTC(),
				Updated: time.Now().UTC(),
			}); err != nil {
				return nil, errors.E(op, errors.Strf("add track failed: %w", err))
			}
		}

		if err := queries.AddTracklistTrack(ctx, db.AddTracklistTrackParams{
			ID:          uuid.NewString(),
			TracklistID: tracklist.ID,
			TrackID:     foundTrackID,
			TrackNumber: int32(idx + 1),
		}); err != nil {
			return nil, errors.E(op, errors.Strf("add tracklist track failed: %w", err))
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, errors.E(op, errors.Strf("tx commit failed: %w", err))
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

func (s *Store) UpdateTracklist(ctx context.Context, id string, model *UpdateTracklistParams) (*Tracklist, error) {
	op := errors.Op("trackliststore[update-tracklist]")

	_, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.E(op, http.StatusNotFound)
	}

	tx, err := s.dataStore.Begin(ctx)
	if err != nil {
		return nil, errors.E(op, errors.Strf("db begin failed: %w", err))
	}
	defer tx.Rollback(ctx)

	queries := s.dataStore.WithTx(tx)

	if _, err = queries.UpdateTracklist(ctx, db.UpdateTracklistParams{
		ID:   id,
		Name: model.Name,
		Date: model.Date,
		URL:  model.URL,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.E(op, http.StatusNotFound)
		}

		return nil, errors.E(op, errors.Strf("update tracklist failed: %w", err))
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, errors.E(op, errors.Strf("tx commit failed: %w", err))
	}

	rows, err := s.dataStore.GetTracklistWithTracks(ctx, id)
	if err != nil {
		return nil, errors.E(op, errors.Strf("find tracklist failed: %w", err))
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

func (s *Store) GetTracklistsByTrack(ctx context.Context, id string, page, limit int32) ([]*Tracklist, int64, error) {
	op := errors.Op("trackliststore[get-tracklists-by-track]")

	var total int64
	total, err := s.dataStore.CountTracklistsByTrack(ctx, id)
	if err != nil {
		return nil, -1, errors.E(op, errors.Strf("count tracklists by track failed: %w", err))
	}

	rows, err := s.dataStore.GetTracklistsByTrack(ctx, db.GetTracklistsByTrackParams{
		TrackID: id,
		Offset:  limit * (page - 1),
		Limit:   limit,
	})
	if err != nil {
		return nil, -1, errors.E(op, errors.Strf("get tracklists by track id failed: %w", err))
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
