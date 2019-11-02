package services

import (
	"fmt"
	"io"
	"math"
	"time"

	"github.com/gofrs/uuid"
	"github.com/tombell/tonality"

	"github.com/tombell/memoir/pkg/datastore"
)

// Tracklist contains data about a specific tracklist. It can contain optional
// track count and list of associated tracks.
type Tracklist struct {
	ID   string    `json:"id"`
	Name string    `json:"name"`
	Date time.Time `json:"date"`

	Created time.Time `json:"-"`
	Updated time.Time `json:"-"`

	TrackCount int      `json:"trackCount"`
	Tracks     []*Track `json:"tracks,omitempty"`
}

// NewTracklist returns a new Tracklist with fields mapped from a database
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

// PagedTracklists contains a paginated list of tracklists, indicating if
// another page is available.
type PagedTracklists struct {
	Tracklists []*Tracklist `json:"tracklists"`
	HasMore    bool         `json:"hasMore"`
}

// NewPagedTracklists returns a new PagedTracklists for the given list of
// tracklists based on the given page and per page amount.
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

// GetTracklists gets all tracklists.
func (s *Services) GetTracklists() ([]*Tracklist, error) {
	tracklists, err := s.DataStore.GetTracklists()
	if err != nil {
		return nil, fmt.Errorf("get tracklists failed: %w", err)
	}

	var models []*Tracklist

	for _, tracklist := range tracklists {
		models = append(models, NewTracklist(tracklist))
	}

	return models, nil
}

// GetTracklist gets a tracklist with the given ID.
func (s *Services) GetTracklist(id string) (*Tracklist, error) {
	tracklist, err := s.DataStore.GetTracklistWithTracks(id)
	if err != nil {
		return nil, fmt.Errorf("get tracklist with tracks failed: %w", err)
	}
	if tracklist == nil {
		return nil, nil
	}

	return NewTracklist(tracklist), nil
}

// GetTracklistByName gets a tracklist with the given name.
func (s *Services) GetTracklistByName(name string) (*Tracklist, error) {
	tracklist, err := s.DataStore.FindTracklistWithTracksByName(name)
	if err != nil {
		return nil, fmt.Errorf("find tracklists with tracks by name failed: %w", err)
	}

	return NewTracklist(tracklist), nil
}

// GetTracklistsByTrack gets all tracklists that include the given track by ID.
func (s *Services) GetTracklistsByTrack(id string) ([]*Tracklist, error) {
	tracklists, err := s.DataStore.FindTracklistsByTrackID(id)
	if err != nil {
		return nil, fmt.Errorf("find tracklists by track id failed: %w", err)
	}

	var models []*Tracklist

	for _, tracklist := range tracklists {
		models = append(models, NewTracklist(tracklist))
	}

	return models, nil
}

// ImportTracklist imports a new tracklist, including any new tracks that have
// not been imported before.
func (s *Services) ImportTracklist(name string, date time.Time, tracks [][]string) (*Tracklist, error) {
	tracklist, err := s.DataStore.FindTracklistByName(name)
	if err != nil {
		return nil, fmt.Errorf("find tracklist failed: %w", err)
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
		return nil, fmt.Errorf("db begin failed: %w", err)
	}

	if err := s.DataStore.AddTracklist(tx, tracklist); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("insert tracklist failed: %w", err)
	}

	for idx, data := range tracks {
		normalizedKey, err := tonality.ConvertKeyToNotation(data[3], tonality.CamelotKeys)
		if err != nil {
			return nil, fmt.Errorf("normalizing key to camelot key notation failed: %w", err)
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
			return nil, fmt.Errorf("import track failed: %w", err)
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
			return nil, fmt.Errorf("insert tracklist_track failed: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("tx commit failed: %w", err)
	}

	return NewTracklist(tracklist), nil
}

// ExportTracklist exports a tracklist with the given name to the given
// io.Writer.
func (s *Services) ExportTracklist(name string, w io.Writer) error {
	tracklist, err := s.DataStore.FindTracklistWithTracksByName(name)
	if err != nil {
		return fmt.Errorf("find tracklist with tracks failed: %w", err)
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

// RemoveTracklist removes a tracklist with the given name.
func (s *Services) RemoveTracklist(name string) error {
	tracklist, err := s.DataStore.FindTracklistByName(name)
	if err != nil {
		return fmt.Errorf("find tracklist failed: %w", err)
	}
	if tracklist == nil {
		return fmt.Errorf("tracklist named %q doesn't exist", name)
	}

	tx, err := s.DataStore.Begin()
	if err != nil {
		return fmt.Errorf("db begin failed: %w", err)
	}

	if err := s.DataStore.RemoveTracklist(tx, tracklist.ID); err != nil {
		tx.Rollback()
		return fmt.Errorf("remove tracklist failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("tx commit failed: %w", err)
	}

	return nil
}
