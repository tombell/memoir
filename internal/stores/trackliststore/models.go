package trackliststore

import (
	"time"

	"github.com/tombell/memoir/internal/stores/trackstore"
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
