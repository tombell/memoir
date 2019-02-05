package database

import "time"

// TracklistRecord represents a single tracklist row in the database.
type TracklistRecord struct {
	ID      string    `db:"id"`
	Name    string    `db:"name"`
	Date    time.Time `db:"date"`
	Created time.Time `db:"created"`
	Updated time.Time `db:"updated"`

	Tracks []*TrackRecord
}

// TrackRecord represents a single track row in the database.
type TrackRecord struct {
	ID      string    `db:"id"`
	Artist  string    `db:"artist"`
	Name    string    `db:"name"`
	Genre   string    `db:"genre"`
	BPM     int       `db:"bpm"`
	Key     string    `db:"key"`
	Created time.Time `db:"created"`
	Updated time.Time `db:"updated"`
}

// TracklistTrackRecord represents a single tracklist_track row in the database.
// Used for mapping a track to a tracklist.
type TracklistTrackRecord struct {
	ID          string `db:"id"`
	TracklistID string `db:"tracklist_id"`
	TrackID     string `db:"track_id"`
	TrackNumber int    `db:"track_number"`
}

// MixUploadRecord represents a single mix upload row in the database.
type MixUploadRecord struct {
	ID          string    `db:"id"`
	TracklistID string    `db:"tracklist_id"`
	Filename    string    `db:"filename"`
	Location    string    `db:"location"`
	Created     time.Time `db:"created"`
	Updated     time.Time `db:"updated"`
}
