package database

import "time"

// TracklistRecord ...
type TracklistRecord struct {
	ID      string
	Created time.Time
	Updated time.Time
}
