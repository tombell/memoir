package services

import (
	"fmt"
	"io"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/tombell/memoir/database"
)

// Tracklist contains data about a specific tracklist.
type Tracklist struct {
	ID      string
	Name    string
	Date    time.Time
	Created time.Time
	Updated time.Time

	Tracks []*Track
}

// NewTracklist returns a new tracklist with fields mapped from a database
// record.
func NewTracklist(record *database.TracklistRecord) *Tracklist {
	tracklist := &Tracklist{
		ID:      record.ID,
		Name:    record.Name,
		Date:    record.Date,
		Created: record.Created,
		Updated: record.Updated,
	}

	for _, track := range record.Tracks {
		tracklist.Tracks = append(tracklist.Tracks, NewTrack(track))
	}

	return tracklist
}

// ImportTracklist imports a new tracklist into the database, including the
// tracklist, and any new tracks that have not been imported before.
func (s *Services) ImportTracklist(name string, date time.Time, tracks [][]string) (*Tracklist, error) {
	tracklist, err := s.DB.FindTracklist(name)
	if err != nil {
		return nil, errors.Wrap(err, "find tracklist failed")
	}
	if tracklist != nil {
		return nil, fmt.Errorf("tracklist named %q already exists", name)
	}

	id, _ := uuid.NewV4()

	tracklist = &database.TracklistRecord{
		ID:      id.String(),
		Name:    name,
		Date:    date,
		Created: time.Now().UTC(),
		Updated: time.Now().UTC(),
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "db begin failed")
	}

	if err := s.DB.InsertTracklist(tx, tracklist); err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "insert tracklist failed")
	}

	for idx, data := range tracks {
		trackImport := &TrackImport{
			Name:   data[0],
			Artist: data[1],
			BPM:    data[2],
			Key:    data[3],
			Genre:  data[4],
		}

		track, err := s.ImportTrack(tx, trackImport)
		if err != nil {
			return nil, errors.Wrap(err, "import track failed")
		}

		id, _ := uuid.NewV4()

		tracklistTrack := &database.TracklistTrackRecord{
			ID:          id.String(),
			TracklistID: tracklist.ID,
			TrackID:     track.ID,
			TrackNumber: idx + 1,
		}

		if err := s.DB.InsertTracklistToTrack(tx, tracklistTrack); err != nil {
			tx.Rollback()
			return nil, errors.Wrap(err, "insert tracklist_track failed")
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "tx commit failed")
	}

	return NewTracklist(tracklist), nil
}

// ExportTracklist exports a tracklist with the given name to the specific
// format.
func (s *Services) ExportTracklist(name string, w io.Writer) error {
	tracklist, err := s.DB.FindTracklistWithTracks(name)
	if err != nil {
		return errors.Wrap(err, "find tracklist with tracks failed")
	}
	if tracklist == nil {
		return fmt.Errorf("tracklist named %q doesn't exist", name)
	}

	for _, track := range tracklist.Tracks {
		str := fmt.Sprintf("%s - %s\n", track.Artist, track.Name)
		w.Write([]byte(str))
	}

	return nil
}

// RemoveTracklist removes a tracklist with the given name from the database.
func (s *Services) RemoveTracklist(name string) error {
	tracklist, err := s.DB.FindTracklist(name)
	if err != nil {
		return errors.Wrap(err, "find tracklist failed")
	}
	if tracklist == nil {
		return fmt.Errorf("tracklist named %q doesn't exist", name)
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return errors.Wrap(err, "db begin failed")
	}

	if err := s.DB.RemoveTracklist(tx, tracklist.ID); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "remove tracklist failed")
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "tx commit failed")
	}

	return nil
}
