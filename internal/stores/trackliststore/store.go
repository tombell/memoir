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

type Store struct{ dataStore *datastore.Store }

func New(store *datastore.Store) *Store { return &Store{dataStore: store} }

func optionalUUID(value string) (*string, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return nil, errors.E("trackliststore[filter]", http.StatusBadRequest)
	}
	normalized := parsed.String()
	return &normalized, nil
}

// GetTracklists uses the same owner and track filters for rows and pagination.
func (s *Store) GetTracklists(ctx context.Context, ownerID, trackID string, page, limit int64) ([]*Tracklist, int64, error) {
	op := errors.Op("trackliststore[get-tracklists]")
	owner, err := optionalUUID(ownerID)
	if err != nil {
		return nil, 0, err
	}
	track, err := optionalUUID(trackID)
	if err != nil {
		return nil, 0, err
	}
	if page < 1 || limit < 1 || limit > 100 || page > 2147483647/limit {
		return nil, 0, errors.E(op, http.StatusBadRequest)
	}
	total, err := s.dataStore.CountTracklists(ctx, db.CountTracklistsParams{OwnerID: owner, TrackID: track})
	if err != nil {
		return nil, 0, errors.E(op, err)
	}
	rows, err := s.dataStore.GetTracklists(ctx, db.GetTracklistsParams{OwnerID: owner, TrackID: track, PageOffset: int32(limit * (page - 1)), PageLimit: int32(limit)})
	if err != nil {
		return nil, 0, errors.E(op, err)
	}
	tracklists := make([]*Tracklist, 0, len(rows))
	for _, row := range rows {
		model := tracklistModel(row.Tracklist, row.OwnerDisplayName)
		model.TrackCount = int(row.TrackCount)
		tracklists = append(tracklists, model)
	}
	return tracklists, total, nil
}

func tracklistModel(row db.Tracklist, ownerName string) *Tracklist {
	return &Tracklist{ID: row.ID, Name: row.Name, Date: row.Date, URL: row.URL, Artwork: row.Artwork,
		Owner: Owner{ID: row.OwnerID, DisplayName: ownerName}, Created: row.Created, Updated: row.Updated}
}

func tracklistWithTracks(rows []*db.GetTracklistWithTracksRow) (*Tracklist, error) {
	if len(rows) == 0 {
		return nil, errors.E("trackliststore[get-tracklist]", http.StatusNotFound)
	}
	model := tracklistModel(rows[0].Tracklist, rows[0].OwnerDisplayName)
	for _, row := range rows {
		model.Tracks = append(model.Tracks, &trackstore.Track{ID: row.Track.ID, Artist: row.Track.Artist, Name: row.Track.Name,
			Genre: row.Track.Genre, BPM: row.Track.BPM, Key: row.Track.Key, Created: row.Track.Created, Updated: row.Track.Updated})
	}
	model.TrackCount = len(model.Tracks)
	return model, nil
}

func (s *Store) GetTracklist(ctx context.Context, id string) (*Tracklist, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.E("trackliststore[get-tracklist]", http.StatusNotFound)
	}
	rows, err := s.dataStore.GetTracklistWithTracks(ctx, id)
	if err != nil {
		return nil, errors.E("trackliststore[get-tracklist]", err)
	}
	return tracklistWithTracks(rows)
}

func writeError(op errors.Op, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.E(op, http.StatusNotFound)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return errors.E(op, http.StatusUnprocessableEntity, errors.M{"name": {"already used by this owner"}})
	}
	return errors.E(op, err)
}

func (s *Store) AddTracklist(ctx context.Context, model *AddTracklistParams) (*Tracklist, error) {
	op := errors.Op("trackliststore[add-tracklist]")
	if _, err := uuid.Parse(model.OwnerID); err != nil {
		return nil, errors.E(op, http.StatusUnauthorized)
	}
	if err := model.Validate(); err != nil {
		return nil, errors.E(op, errors.M(err), http.StatusUnprocessableEntity)
	}
	tx, err := s.dataStore.Begin(ctx)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer tx.Rollback(ctx)
	q := s.dataStore.WithTx(tx)
	tracklist, err := q.AddTracklist(ctx, model.ToDatabaseParams())
	if err != nil {
		return nil, writeError(op, err)
	}
	for idx, data := range model.Tracks {
		row, err := q.GetTrackByArtistAndName(ctx, db.GetTrackByArtistAndNameParams{Artist: data[1], Name: data[0]})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.E(op, err)
		}
		var trackID string
		if errors.Is(err, pgx.ErrNoRows) {
			trackID = uuid.NewString()
			bpm, _ := strconv.ParseFloat(data[2], 64)
			now := time.Now().UTC()
			if err := q.AddTrack(ctx, db.AddTrackParams{ID: trackID, Name: data[0], Artist: data[1], BPM: bpm,
				Key: strings.ToUpper(data[3]), Genre: data[4], Created: now, Updated: now}); err != nil {
				return nil, errors.E(op, err)
			}
		} else {
			trackID = row.ID
		}
		if err := q.AddTracklistTrack(ctx, db.AddTracklistTrackParams{ID: uuid.NewString(), TracklistID: tracklist.ID, TrackID: trackID, TrackNumber: int32(idx + 1)}); err != nil {
			return nil, errors.E(op, err)
		}
	}
	rows, err := q.GetTracklistWithTracks(ctx, tracklist.ID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	result, err := tracklistWithTracks(rows)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, errors.E(op, err)
	}
	return result, nil
}

func (s *Store) UpdateTracklist(ctx context.Context, id, ownerID string, model *UpdateTracklistParams) (*Tracklist, error) {
	op := errors.Op("trackliststore[update-tracklist]")
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.E(op, http.StatusNotFound)
	}
	if _, err := uuid.Parse(ownerID); err != nil {
		return nil, errors.E(op, http.StatusUnauthorized)
	}
	if err := model.Validate(); err != nil {
		return nil, errors.E(op, errors.M(err), http.StatusUnprocessableEntity)
	}
	tx, err := s.dataStore.Begin(ctx)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer tx.Rollback(ctx)
	q := s.dataStore.WithTx(tx)
	if _, err := q.UpdateTracklist(ctx, model.ToDatabaseParams(id, ownerID)); err != nil {
		return nil, writeError(op, err)
	}
	rows, err := q.GetTracklistWithTracks(ctx, id)
	if err != nil {
		return nil, errors.E(op, err)
	}
	result, err := tracklistWithTracks(rows)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, errors.E(op, err)
	}
	return result, nil
}

// DeleteTracklist checks and locks the owner row before touching associations.
func (s *Store) DeleteTracklist(ctx context.Context, id, ownerID string) error {
	op := errors.Op("trackliststore[delete-tracklist]")
	if _, err := uuid.Parse(id); err != nil {
		return errors.E(op, http.StatusNotFound)
	}
	if _, err := uuid.Parse(ownerID); err != nil {
		return errors.E(op, http.StatusUnauthorized)
	}
	tx, err := s.dataStore.Begin(ctx)
	if err != nil {
		return errors.E(op, err)
	}
	defer tx.Rollback(ctx)
	q := s.dataStore.WithTx(tx)
	if _, err := q.LockOwnedTracklist(ctx, db.LockOwnedTracklistParams{ID: id, OwnerID: ownerID}); err != nil {
		return writeError(op, err)
	}
	if err := q.DeleteTracklistTracks(ctx, id); err != nil {
		return errors.E(op, err)
	}
	count, err := q.DeleteTracklist(ctx, db.DeleteTracklistParams{ID: id, OwnerID: ownerID})
	if err != nil {
		return errors.E(op, err)
	}
	if count != 1 {
		return errors.E(op, http.StatusNotFound)
	}
	if err := tx.Commit(ctx); err != nil {
		return errors.E(op, err)
	}
	return nil
}
