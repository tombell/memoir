package services

import (
	"fmt"
	"strconv"
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
}

// NewTracklist returns a new tracklist with fields mapped from a database
// record.
func NewTracklist(tracklist *database.TracklistRecord) *Tracklist {
	return &Tracklist{
		ID:      tracklist.ID,
		Name:    tracklist.Name,
		Date:    tracklist.Date,
		Created: tracklist.Created,
		Updated: tracklist.Updated,
	}
}

// ImportTracklist imports a new tracklist into the database, including the
// tracklist, and any new tracks that have not been imported before.
// TODO: refactor into smaller chunks?
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

	var records []*database.TrackRecord

	for _, data := range tracks {
		track, err := s.DB.FindTrack(data[1], data[0])
		if err != nil {
			tx.Rollback()
			return nil, errors.Wrap(err, "find track failed")
		}
		if track != nil {
			records = append(records, track)
			continue
		}

		id, _ := uuid.NewV4()
		bpm, _ := strconv.Atoi(data[2])

		track = &database.TrackRecord{
			ID:      id.String(),
			Name:    data[0],
			Artist:  data[1],
			BPM:     bpm,
			Key:     data[3],
			Genre:   data[4],
			Created: time.Now().UTC(),
			Updated: time.Now().UTC(),
		}

		if err := s.DB.InsertTrack(tx, track); err != nil {
			tx.Rollback()
			return nil, errors.Wrap(err, "insert track failed")
		}

		records = append(records, track)
	}

	for idx, track := range records {
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
