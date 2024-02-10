package services

import (
	"time"

	"github.com/charmbracelet/log"

	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/datastore"
	"github.com/tombell/memoir/internal/filestore"
)

type Services struct {
	Config    *config.Config
	DataStore *datastore.Store
	FileStore *filestore.Store
	Logger    *log.Logger
}

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

type Track struct {
	ID     string  `json:"id"`
	Artist string  `json:"artist"`
	Name   string  `json:"name"`
	Genre  string  `json:"genre"`
	BPM    float64 `json:"bpm"`
	Key    string  `json:"key"`

	Created time.Time `json:"-"`
	Updated time.Time `json:"-"`

	Played int64 `json:"played,omitempty"`

	ArtistHighlighted string `json:"artistHighlighted,omitempty"`
	NameHighlighted   string `json:"nameHighlighted,omitempty"`
}

type UploadedItem struct {
	Key string `json:"key"`
}
