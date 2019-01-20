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
	s.Logger.Printf("checking if tracklist %q already exists...\n", name)

	tracklist, err := s.DB.FindTracklist(name)
	if err != nil {
		return nil, err
	}
	if tracklist != nil {
		tl := &Tracklist{
			ID:      tracklist.ID,
			Name:    tracklist.Name,
			Date:    tracklist.Date,
			Created: tracklist.Created,
			Updated: tracklist.Updated,
		}

		return tl, fmt.Errorf("tracklist named %q already exists", name)
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

	s.Logger.Println("tracklist doesn't exist, creating tracklist...")

	if err := s.DB.InsertTracklist(tx, tracklist); err != nil {
		tx.Rollback()
		return nil, err
	}

	var records []*database.TrackRecord

	for _, data := range tracks {
		s.Logger.Printf("checking if track \"%s - %s\" already exists...\n", data[1], data[0])

		track, err := s.DB.FindTrack(data[1], data[0])
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		if track != nil {
			s.Logger.Printf("track already exists: %q\n", track.ID)

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

		s.Logger.Printf("track doesn't exist, creating track...\n")

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
