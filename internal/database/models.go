// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package database

import (
	"time"
)

type Track struct {
	ID               string
	Artist           string
	Name             string
	Genre            string
	BPM              float64
	Key              string
	Created          time.Time
	Updated          time.Time
	FtsNameAndArtist string
}

type Tracklist struct {
	ID      string
	Name    string
	Date    time.Time
	Artwork string
	URL     string
	Created time.Time
	Updated time.Time
}
