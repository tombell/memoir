package trackstore

import "time"

// Track is the model used for serialising a tracklist to JSON.
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
