package trackliststore

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/tombell/valid"

	db "github.com/tombell/memoir/internal/database"
)

// PatchField distinguishes omitted JSON fields from supplied values and null.
// Invalid values are retained as validation errors, rather than zero values.
type PatchField[T any] struct {
	Value   T
	Present bool
	Null    bool
	Invalid bool
}

func (f *PatchField[T]) UnmarshalJSON(data []byte) error {
	*f = PatchField[T]{Present: true}
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		f.Null = true
		return nil
	}
	f.Invalid = json.Unmarshal(data, &f.Value) != nil
	return nil
}

// UpdateTracklistParams contains only the metadata supplied by PATCH.
// Omitted fields retain their values. Null and empty values are rejected.
type UpdateTracklistParams struct {
	Name    PatchField[string]    `json:"name"`
	Date    PatchField[string]    `json:"date"`
	URL     PatchField[string]    `json:"url"`
	Artwork PatchField[string]    `json:"artwork"`
	Tracks  PatchField[TrackRows] `json:"tracks"`
}

func (t *UpdateTracklistParams) Validate() valid.Error {
	v := valid.New()
	check := func(name string, field PatchField[string]) bool {
		if !field.Present {
			return false
		}
		v.Check(name,
			valid.Case{Cond: !field.Null, Msg: "Must not be null"},
			valid.Case{Cond: !field.Invalid, Msg: "Must be a string"},
		)
		if field.Null || field.Invalid {
			return false
		}
		v.Check(name,
			valid.Case{Cond: strings.TrimSpace(field.Value) != "", Msg: "Must not be empty"},
			valid.Case{Cond: !strings.ContainsRune(field.Value, 0), Msg: "Must not contain a null character"},
			valid.Case{Cond: valid.MaxLength(field.Value, 256), Msg: "Must be less than, or equal to 256 characters"},
		)
		return true
	}
	check("name", t.Name)
	if check("date", t.Date) {
		v.Check("date", valid.Case{Cond: valid.IsDate(t.Date.Value), Msg: "Must be a valid ISO 8601 date"})
	}
	if check("url", t.URL) {
		v.Check("url", valid.Case{Cond: valid.IsURL(t.URL.Value), Msg: "Must be a valid URL"})
	}
	check("artwork", t.Artwork)
	if t.Tracks.Present {
		v.Check("tracks",
			valid.Case{Cond: !t.Tracks.Null, Msg: "Must not be null"},
			valid.Case{Cond: !t.Tracks.Invalid, Msg: "Must be an array of string arrays"},
		)
		if !t.Tracks.Null && !t.Tracks.Invalid {
			validateReplacementTracks(v, t.Tracks.Value)
		}
	}
	if v.Valid() {
		return nil
	}
	return v.Errors
}

func (t *UpdateTracklistParams) ToDatabaseParams(id string) db.UpdateTracklistParams {
	date, _ := time.Parse(time.RFC3339, t.Date.Value)
	return db.UpdateTracklistParams{
		ID:   id,
		Name: t.Name.Value, UpdateName: t.Name.Present,
		Date: date, UpdateDate: t.Date.Present,
		URL: t.URL.Value, UpdateUrl: t.URL.Present,
		Artwork: t.Artwork.Value, UpdateArtwork: t.Artwork.Present,
	}
}
