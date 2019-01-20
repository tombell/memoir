package services

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofrs/uuid"

	"github.com/tombell/memoir/database"
)

// ImportTracklist imports a new tracklist into the database, including the
// tracklist, and any new tracks that have not been imported before.
func (s *Services) ImportTracklist(name string, date time.Time, tracks [][]string) error {
	tracklist, err := s.DB.FindTracklist(name)
	if err != nil {
		return err
	}
	if tracklist != nil {
		return fmt.Errorf("tracklist named %q already exists", name)
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
		return err
	}

	if err := s.DB.InsertTracklist(tx, tracklist); err != nil {
		tx.Rollback()
		return err
	}

	var records []*database.TrackRecord

	for _, data := range tracks {
		track, err := s.DB.FindTrack(data[1], data[0])
		if err != nil {
			tx.Rollback()
			return err
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
			return err
		}

		records = append(records, track)
	}

	for _, track := range records {
		if err := s.DB.InsertTracklistToTrack(tx, tracklist.ID, track.ID); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
