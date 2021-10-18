package services

import (
	"time"

	"github.com/tombell/memoir/internal/jsondate"
	"github.com/tombell/memoir/pkg/datastore"
)

// TracklistAdd contains data about a tracklist to add.
type TracklistAdd struct {
	Name    string        `json:"name"`
	Date    jsondate.Date `json:"date"`
	URL     string        `json:"url"`
	Artwork string        `json:"artwork"`
	Tracks  [][]string    `json:"tracks"`
}

// TracklistUpdate contains data about a tracklist to update.
type TracklistUpdate struct {
	Name string        `json:"name"`
	Date jsondate.Date `json:"date"`
	URL  string        `json:"url"`
}

// Tracklist contains data about a specific tracklist. It can contain optional
// track count and list of associated tracks.
type Tracklist struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Date    jsondate.Date `json:"date"`
	URL     string        `json:"url"`
	Artwork string        `json:"artwork"`

	Created time.Time `json:"-"`
	Updated time.Time `json:"-"`

	TrackCount int      `json:"trackCount"`
	Tracks     []*Track `json:"tracks,omitempty"`
}

// NewTracklist returns a new Tracklist with fields mapped from a database
// record.
func NewTracklist(record *datastore.Tracklist) *Tracklist {
	tracklist := &Tracklist{
		ID:         record.ID,
		Name:       record.Name,
		Date:       jsondate.Date{Time: record.Date},
		URL:        record.URL,
		Artwork:    record.Artwork,
		Created:    record.Created,
		Updated:    record.Updated,
		TrackCount: record.TrackCount,
	}

	if len(record.Tracks) > 0 {
		tracklist.TrackCount = len(record.Tracks)
	}

	for _, track := range record.Tracks {
		tracklist.Tracks = append(tracklist.Tracks, NewTrack(track))
	}

	return tracklist
}
