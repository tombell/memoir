package services

import (
	"time"

	"github.com/tombell/memoir/internal/datastore"
)

type Tracklist struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Date    time.Time `json:"date"`
	URL     string    `json:"url"`
	Artwork string    `json:"artwork"`

	Created time.Time `json:"-"`
	Updated time.Time `json:"-"`

	Tracks     []*Track `json:"tracks,omitempty"`
	TrackCount int      `json:"trackCount"`
}

type TracklistAdd struct {
	Name    string     `json:"name"`
	Date    time.Time  `json:"date"`
	URL     string     `json:"url"`
	Artwork string     `json:"artwork"`
	Tracks  [][]string `json:"tracks"`
}

type TracklistUpdate struct {
	Name string    `json:"name"`
	Date time.Time `json:"date"`
	URL  string    `json:"url"`
}

func NewTracklist(record *datastore.Tracklist) *Tracklist {
	tracklist := &Tracklist{
		ID:         record.ID,
		Name:       record.Name,
		Date:       record.Date,
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
