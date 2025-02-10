package trackliststore

import (
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

	Created time.Time `json:"-"`
	Updated time.Time `json:"-"`

	Tracks     []*trackstore.Track `json:"tracks,omitempty"`
	TrackCount int                 `json:"trackCount"`
}

// AddTracklistParams are the parameters deserialised from JSON for adding a new
// tracklist.
type AddTracklistParams struct {
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
		valid.Case{Cond: valid.NotEmpty(t.URL), Msg: "Must not be empty"},
		valid.Case{Cond: valid.MaxLength(t.URL, 256), Msg: "Must be less than, or equal to 256 characters"},
	)
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

	if validator.Valid() {
		return nil
	}

	return validator.Errors
}

// ToDatabaseParams returns a database params struct for updating an existing
// tracklist.
func (t *UpdateTracklistParams) ToDatabaseParams(id string) db.UpdateTracklistParams {
	date, _ := time.Parse(time.RFC3339, t.Date)

	return db.UpdateTracklistParams{
		ID:   id,
		Name: t.Name,
		Date: date,
		URL:  t.URL,
	}
}
