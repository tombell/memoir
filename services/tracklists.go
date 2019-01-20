package services

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofrs/uuid"

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

// ImportTracklist imports a new tracklist into the database, including the
// tracklist, and any new tracks that have not been imported before.
func (s *Services) ImportTracklist(name string, date time.Time, tracks [][]string) (*Tracklist, error) {
	tracklist, err := s.DB.FindTracklist(name)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	if err := s.DB.InsertTracklist(tx, tracklist); err != nil {
		tx.Rollback()
		return nil, err
	}

	var records []*database.TrackRecord

	for _, data := range tracks {
		track, err := s.DB.FindTrack(data[1], data[0])
		if err != nil {
			tx.Rollback()
			return nil, err
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
			return nil, err
		}

		records = append(records, track)
	}

	for _, track := range records {
		if err := s.DB.InsertTracklistToTrack(tx, tracklist.ID, track.ID); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, err
	}

	tl := &Tracklist{
		ID:      tracklist.ID,
		Name:    tracklist.Name,
		Date:    tracklist.Date,
		Created: tracklist.Created,
		Updated: tracklist.Updated,
	}

	return tl, nil
}

// RemoveTracklist removes a tracklist with the given name from the database.
func (s *Services) RemoveTracklist(name string) error {
	tracklist, err := s.DB.FindTracklist(name)
	if err != nil {
		return err
	}
	if tracklist == nil {
		return fmt.Errorf("tracklist named %q doesn't exist", name)
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	if err := s.DB.RemoveTracklist(tx, tracklist.ID); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
