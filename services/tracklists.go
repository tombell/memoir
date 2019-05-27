package services

import (
	"fmt"
	"io"
	"math"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/tombell/tonality"

	"github.com/tombell/memoir/datastore"
)

// Tracklist contains data about a specific tracklist.
type Tracklist struct {
	ID   string    `json:"id"`
	Name string    `json:"name"`
	Date time.Time `json:"date"`

	Created time.Time `json:"-"`
	Updated time.Time `json:"-"`

	TrackCount int      `json:"trackCount"`
	Tracks     []*Track `json:"tracks,omitempty"`
}

// NewTracklist returns a new tracklist with fields mapped from a database
// record.
func NewTracklist(record *datastore.Tracklist) *Tracklist {
	tracklist := &Tracklist{
		ID:         record.ID,
		Name:       record.Name,
		Date:       record.Date,
		Created:    record.Created,
		Updated:    record.Updated,
		TrackCount: record.TrackCount,
	}

	if len(record.Tracks) > 0 {
		tracklist.TrackCount = len(record.Tracks)
	}

	for _, track := range record.Tracks {
		tracklist.Tracks = append(tracklist.Tracks, NewTrack(track))
	}

	return tracklist
}

// PagedTracklists ...
type PagedTracklists struct {
	Tracklists []*Tracklist `json:"tracklists"`
	HasMore    bool         `json:"hasMore"`
}

// NewPagedTracklists ...
func NewPagedTracklists(tracklists []*Tracklist, page, perPage int) *PagedTracklists {
	total := len(tracklists)
	pages := math.Ceil(float64(total) / float64(perPage))
	offset := (page - 1) * perPage
	count := offset + perPage

	if count > total {
		count = total
	}

	return &PagedTracklists{
		Tracklists: tracklists[offset:count],
		HasMore:    page < int(pages),
	}
}

// GetTracklists ...
func (s *Services) GetTracklists() ([]*Tracklist, error) {
	tracklists, err := s.DataStore.GetTracklists()
	if err != nil {
		return nil, errors.Wrap(err, "get tracklists failed")
	}

	var models []*Tracklist

	for _, tracklist := range tracklists {
		models = append(models, NewTracklist(tracklist))
	}

	return models, nil
}

// GetTracklist ...
func (s *Services) GetTracklist(id string) (*Tracklist, error) {
	tracklist, err := s.DataStore.GetTracklistWithTracks(id)
	if err != nil {
		return nil, errors.Wrap(err, "get tracklist with tracks failed")
	}
	if tracklist == nil {
		return nil, nil
	}

	return NewTracklist(tracklist), nil
}

// GetTracklistByName ...
func (s *Services) GetTracklistByName(name string) (*Tracklist, error) {
	tracklist, err := s.DataStore.FindTracklistWithTracksByName(name)
	if err != nil {
		return nil, errors.Wrap(err, "find tracklists with tracks by name failed")
	}

	return NewTracklist(tracklist), nil
}

// GetTracklistsByTrack ...
func (s *Services) GetTracklistsByTrack(id string) ([]*Tracklist, error) {
	tracklists, err := s.DataStore.FindTracklistsByTrackID(id)
	if err != nil {
		return nil, errors.Wrap(err, "find tracklists by track id failed")
	}

	var models []*Tracklist

	for _, tracklist := range tracklists {
		models = append(models, NewTracklist(tracklist))
	}

	return models, nil
}

// ImportTracklist imports a new tracklist into the database, including the
// tracklist, and any new tracks that have not been imported before.
func (s *Services) ImportTracklist(name string, date time.Time, tracks [][]string) (*Tracklist, error) {
	tracklist, err := s.DataStore.FindTracklistByName(name)
	if err != nil {
		return nil, errors.Wrap(err, "find tracklist failed")
	}
	if tracklist != nil {
		return nil, fmt.Errorf("tracklist named %q already exists", name)
	}

	id, _ := uuid.NewV4()

	tracklist = &datastore.Tracklist{
		ID:      id.String(),
		Name:    name,
		Date:    date,
		Created: time.Now().UTC(),
		Updated: time.Now().UTC(),
	}

	tx, err := s.DataStore.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "db begin failed")
	}

	if err := s.DataStore.AddTracklist(tx, tracklist); err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "insert tracklist failed")
	}

	for idx, data := range tracks {
		normalizedKey, err := tonality.ConvertKeyToNotation(data[3], tonality.CamelotKeys)
		if err != nil {
			return nil, errors.Wrap(err, "normalizing key to camelot key notation failed")
		}

		trackImport := &TrackImport{
			Name:   data[0],
			Artist: data[1],
			BPM:    data[2],
			Key:    normalizedKey,
			Genre:  data[4],
		}

		track, err := s.ImportTrack(tx, trackImport)
		if err != nil {
			return nil, errors.Wrap(err, "import track failed")
		}

		id, _ := uuid.NewV4()

		tracklistTrack := &datastore.TracklistTrack{
			ID:          id.String(),
			TracklistID: tracklist.ID,
			TrackID:     track.ID,
			TrackNumber: idx + 1,
		}

		if err := s.DataStore.AddTracklistTrack(tx, tracklistTrack); err != nil {
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
	tracklist, err := s.DataStore.FindTracklistWithTracksByName(name)
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
	tracklist, err := s.DataStore.FindTracklistByName(name)
	if err != nil {
		return errors.Wrap(err, "find tracklist failed")
	}
	if tracklist == nil {
		return fmt.Errorf("tracklist named %q doesn't exist", name)
	}

	tx, err := s.DataStore.Begin()
	if err != nil {
		return errors.Wrap(err, "db begin failed")
	}

	if err := s.DataStore.RemoveTracklist(tx, tracklist.ID); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "remove tracklist failed")
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "tx commit failed")
	}

	return nil
}
