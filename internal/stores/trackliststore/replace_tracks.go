package trackliststore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tombell/valid"

	db "github.com/tombell/memoir/internal/database"
	"github.com/tombell/memoir/internal/errors"
)

// TrackRows preserves the existing [name, artist, bpm, key, genre] format.
type TrackRows [][]string

// UnmarshalJSON rejects null cells, which encoding/json otherwise turns into
// empty strings. Row lengths and values are checked before starting an update.
func (r *TrackRows) UnmarshalJSON(data []byte) error {
	var raw [][]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	rows := make(TrackRows, len(raw))
	for i, row := range raw {
		rows[i] = make([]string, len(row))
		for j, cell := range row {
			if bytes.Equal(bytes.TrimSpace(cell), []byte("null")) {
				return fmt.Errorf("tracks[%d][%d] must be a string", i, j)
			}
			if err := json.Unmarshal(cell, &rows[i][j]); err != nil {
				return err
			}
		}
	}
	*r = rows
	return nil
}

func validateReplacementTracks(v *valid.Validator, tracks TrackRows) {
	v.Check("tracks", valid.Case{Cond: len(tracks) > 0, Msg: "Must not be empty"})
	v.Check("tracks", valid.Case{Cond: int64(len(tracks)) <= math.MaxInt32, Msg: "Too many tracks"})
	for i, row := range tracks {
		field := fmt.Sprintf("tracks[%d]", i)
		v.Check(field, valid.Case{Cond: len(row) == 5, Msg: "Must contain five fields: name, artist, bpm, key, genre"})
		if len(row) != 5 {
			continue
		}
		for j, spec := range []struct {
			name     string
			max      int
			required bool
		}{
			{"name", 256, true}, {"artist", 256, true}, {"bpm", 0, false}, {"key", 8, false}, {"genre", 128, false},
		} {
			if j == 2 {
				continue
			}
			v.Check(field+"."+spec.name,
				valid.Case{Cond: !spec.required || valid.NotEmpty(row[j]), Msg: "Must not be empty"},
				valid.Case{Cond: valid.MaxLength(row[j], spec.max), Msg: fmt.Sprintf("Must be less than, or equal to %d characters", spec.max)},
				valid.Case{Cond: !strings.ContainsRune(row[j], 0), Msg: "Must not contain a null character"},
			)
		}
		bpm, err := strconv.ParseFloat(row[2], 64)
		v.Check(field+".bpm", valid.Case{Cond: err == nil && !math.IsNaN(bpm) && !math.IsInf(bpm, 0), Msg: "Must be a finite number"})
	}
}

// replaceTracklistTracks runs entirely in the caller's transaction. It only
// deletes associations; existing track records and their metadata are retained.
func replaceTracklistTracks(ctx context.Context, queries *db.Queries, id string, tracks TrackRows) error {
	// Serialize PATCH identity reuse while the baseline schema has no unique
	// artist/name constraint. Creation's identity constraint is a separate change.
	if err := queries.LockPatchTrackIdentity(ctx); err != nil {
		return errors.Strf("lock track identity failed: %w", err)
	}
	if err := queries.DeleteTracklistTracks(ctx, id); err != nil {
		return errors.Strf("delete tracklist associations failed: %w", err)
	}
	for i, data := range tracks {
		row, err := queries.GetTrackByArtistAndName(ctx, db.GetTrackByArtistAndNameParams{Artist: data[1], Name: data[0]})
		var trackID string
		switch {
		case err == nil:
			trackID = row.ID
		case errors.Is(err, pgx.ErrNoRows):
			trackID = uuid.NewString()
			bpm, err := strconv.ParseFloat(data[2], 64)
			if err != nil {
				return err
			}
			now := time.Now().UTC()
			if err := queries.AddTrack(ctx, db.AddTrackParams{
				ID: trackID, Name: data[0], Artist: data[1], BPM: bpm,
				Key: strings.ToUpper(data[3]), Genre: data[4], Created: now, Updated: now,
			}); err != nil {
				return errors.Strf("add replacement track failed: %w", err)
			}
		default:
			return errors.Strf("find replacement track failed: %w", err)
		}
		if err := queries.AddTracklistTrack(ctx, db.AddTracklistTrackParams{
			ID: uuid.NewString(), TracklistID: id, TrackID: trackID, TrackNumber: int32(i + 1),
		}); err != nil {
			return errors.Strf("add replacement association failed: %w", err)
		}
	}
	return nil
}
