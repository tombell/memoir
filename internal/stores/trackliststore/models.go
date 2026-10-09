package trackliststore

import (
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/tombell/valid"

	db "github.com/tombell/memoir/internal/database"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

// Tracklist is the model used for serialising a tracklist to JSON.
type Tracklist struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Date    time.Time `json:"date"`
	URL     string    `json:"url"`
	Artwork string    `json:"artwork"`
	Owner   Owner     `json:"owner"`

	Created time.Time `json:"-"`
	Updated time.Time `json:"-"`

	Tracks     []*trackstore.Track `json:"tracks,omitempty"`
	TrackCount int                 `json:"trackCount"`
}

type Owner struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// AddTracklistParams are the parameters deserialised from JSON for adding a new
// tracklist.
type AddTracklistParams struct {
	OwnerID string     `json:"-"`
	Name    string     `json:"name"`
	Date    string     `json:"date"`
	URL     string     `json:"url"`
	Artwork string     `json:"artwork"`
	Tracks  [][]string `json:"tracks"`
}

// Validate validate that the data provided is correct for adding a new
// tracklist.
func (t *AddTracklistParams) Validate() valid.Error {
	validator := valid.New()
	validator.Check("name",
		valid.Case{Cond: valid.NotEmpty(t.Name), Msg: "Must not be empty"},
		valid.Case{Cond: valid.MaxLength(t.Name, 256), Msg: "Must be less than, or equal to 256 characters"},
	)
	validator.Check("date",
		valid.Case{Cond: valid.NotEmpty(t.Date), Msg: "Must not be empty"},
		valid.Case{Cond: valid.IsDate(t.Date), Msg: "Must be a valid ISO 8601 date"},
	)
	validator.Check("url",
		valid.Case{Cond: valid.NotEmpty(t.URL), Msg: "Must not be empty"},
		valid.Case{Cond: valid.MaxLength(t.URL, 256), Msg: "Must be less than, or equal to 256 characters"},
		valid.Case{Cond: valid.IsURL(t.URL), Msg: "Must be a valid URL"},
	)
	validator.Check("artwork",
		valid.Case{Cond: valid.NotEmpty(t.Artwork), Msg: "Must not be empty"},
		valid.Case{Cond: valid.MaxLength(t.Artwork, 256), Msg: "Must be less than, or equal to 256 characters"},
	)
	for _, track := range t.Tracks {
		if len(track) != 5 {
			validator.Check("tracks", valid.Case{Cond: false, Msg: "Each track must contain name, artist, BPM, key, and genre"})
			continue
		}
		bpm, err := strconv.ParseFloat(track[2], 64)
		validator.Check("tracks",
			valid.Case{Cond: track[0] != "" && valid.MaxLength(track[0], 256) && track[1] != "" && valid.MaxLength(track[1], 256), Msg: "Track name and artist must contain 1 to 256 characters"},
			valid.Case{Cond: err == nil && !math.IsNaN(bpm) && !math.IsInf(bpm, 0) && bpm >= 0, Msg: "BPM must be a finite nonnegative number"},
			valid.Case{Cond: valid.MaxLength(track[3], 8) && valid.MaxLength(track[4], 128), Msg: "Track key or genre is too long"},
		)
	}
	_, dateErr := time.Parse(time.RFC3339, t.Date)
	validator.Check("date", valid.Case{Cond: dateErr == nil, Msg: "Must be an RFC 3339 timestamp"})
	validator.Check("tracks",
		valid.Case{Cond: len(t.Tracks) != 0, Msg: "Must not be empty"},
	)

	if validator.Valid() {
		return nil
	}

	return validator.Errors
}

// ToDatabaseParams returns a database params struct for adding a new tracklist.
func (t *AddTracklistParams) ToDatabaseParams() db.AddTracklistParams {
	date, _ := time.Parse(time.RFC3339, t.Date)

	return db.AddTracklistParams{
		ID:      uuid.NewString(),
		Name:    t.Name,
		Date:    date,
		URL:     t.URL,
		Artwork: t.Artwork,
		OwnerID: t.OwnerID,
	}
}

// UpdateTracklistParams are the parameters deserialised from JSON for updating
// an existing tracklist.
type UpdateTracklistParams struct {
	Name string `json:"name"`
	Date string `json:"date"`
	URL  string `json:"url"`
}

// Validate validates that the data provided is correct for updating an existing
// tracklist.
func (t *UpdateTracklistParams) Validate() valid.Error {
	validator := valid.New()
	validator.Check("name",
		valid.Case{Cond: valid.NotEmpty(t.Name), Msg: "Must not be empty"},
		valid.Case{Cond: valid.MaxLength(t.Name, 256), Msg: "Must be less than, or equal to 256 characters"},
	)
	validator.Check("date",
		valid.Case{Cond: valid.NotEmpty(t.Date), Msg: "Must not be empty"},
		valid.Case{Cond: valid.IsDate(t.Date), Msg: "Must be a valid ISO 8601 date"},
	)
	validator.Check("url",
		valid.Case{Cond: valid.NotEmpty(t.URL), Msg: "Must not be empty"},
		valid.Case{Cond: valid.MaxLength(t.URL, 256), Msg: "Must be less than, or equal to 256 characters"},
		valid.Case{Cond: valid.IsURL(t.URL), Msg: "Must be a valid URL"},
	)
	_, dateErr := time.Parse(time.RFC3339, t.Date)
	validator.Check("date", valid.Case{Cond: dateErr == nil, Msg: "Must be an RFC 3339 timestamp"})

	if validator.Valid() {
		return nil
	}

	return validator.Errors
}

// ToDatabaseParams returns a database params struct for updating an existing
// tracklist.
func (t *UpdateTracklistParams) ToDatabaseParams(id, ownerID string) db.UpdateTracklistParams {
	date, _ := time.Parse(time.RFC3339, t.Date)

	return db.UpdateTracklistParams{
		ID:      id,
		Name:    t.Name,
		Date:    date,
		URL:     t.URL,
		OwnerID: ownerID,
	}
}
