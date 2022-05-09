package datastore

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/tombell/memoir/internal/datastore/queries"
)

type Tracklist struct {
	ID      string
	Name    string
	Artwork string
	URL     string
	Date    time.Time
	Created time.Time
	Updated time.Time

	TrackCount int
	Tracks     []*Track
}

func (s *Store) AddTracklist(tx *sql.Tx, tracklist *Tracklist) error {
	if _, err := tx.Exec(
		queries.AddTracklist,
		tracklist.ID,
		tracklist.Name,
		tracklist.URL,
		tracklist.Artwork,
		tracklist.Date,
		tracklist.Created,
		tracklist.Updated,
	); err != nil {
		return fmt.Errorf("tx exec failed: %w", err)
	}

	return nil
}

func (s *Store) UpdateTracklist(tx *sql.Tx, id, name, url string, date time.Time) error {
	if _, err := tx.Exec(
		queries.UpdateTracklist,
		id,
		name,
		url,
		date,
	); err != nil {
		return fmt.Errorf("tx exec failed: %w", err)
	}

	return nil
}

func (s *Store) GetTracklistsCount() (int, error) {
	var count struct {
		Count int
	}

	if err := s.DB.Get(&count, queries.GetTracklistsCount); err != nil {
		return -1, fmt.Errorf("db get failed: %w", err)
	}

	return count.Count, nil
}

func (s *Store) GetTracklists(offset, limit int) ([]*Tracklist, error) {
	rows, err := s.Queryx(queries.GetTracklists, offset, limit)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}

	var tracklists []*Tracklist

	for rows.Next() {
		var tracklist Tracklist

		if err := rows.StructScan(&tracklist); err != nil {
			return nil, fmt.Errorf("rows scan failed: %w", err)
		}

		tracklists = append(tracklists, &tracklist)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows next failed: %w", err)
	}

	return tracklists, nil
}

func (s *Store) FindTracklist(id string) (*Tracklist, error) {
	var tracklist Tracklist

	switch err := s.QueryRowx(queries.FindTracklistByID, id).StructScan(&tracklist); err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &tracklist, nil
	default:
		return nil, fmt.Errorf("db query row failed: %w", err)
	}
}

func (s *Store) FindTracklistWithTracks(id string) (*Tracklist, error) {
	rows, err := s.Queryx(queries.FindTracklistWithTracksByID, id)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}

	var tracklist Tracklist

	for rows.Next() {
		var track Track

		if err := rows.Scan(
			&tracklist.ID,
			&tracklist.Name,
			&tracklist.Date,
			&tracklist.Artwork,
			&tracklist.URL,
			&tracklist.Created,
			&tracklist.Updated,
			&track.ID,
			&track.Artist,
			&track.Name,
			&track.Genre,
			&track.BPM,
			&track.Key,
			&track.Created,
			&track.Updated,
		); err != nil {
			return nil, fmt.Errorf("rows scan failed: %w", err)
		}

		tracklist.Tracks = append(tracklist.Tracks, &track)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows next failed: %w", err)
	}

	if tracklist.ID == "" {
		return nil, nil
	}

	return &tracklist, nil
}

func (s *Store) FindTracklistByName(name string) (*Tracklist, error) {
	var tracklist Tracklist

	switch err := s.QueryRowx(queries.FindTracklistByName, name).StructScan(&tracklist); err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &tracklist, nil
	default:
		return nil, fmt.Errorf("db query row failed: %w", err)
	}
}

func (s *Store) FindTracklistWithTracksByName(name string) (*Tracklist, error) {
	rows, err := s.Queryx(queries.FindTracklistWithTracksByName, name)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}

	var tracklist Tracklist

	for rows.Next() {
		var track Track

		if err := rows.Scan(
			&tracklist.ID,
			&tracklist.Name,
			&tracklist.Date,
			&tracklist.Artwork,
			&tracklist.URL,
			&tracklist.Created,
			&tracklist.Updated,
			&track.ID,
			&track.Artist,
			&track.Name,
			&track.Genre,
			&track.BPM,
			&track.Key,
			&track.Created,
			&track.Updated,
		); err != nil {
			return nil, fmt.Errorf("rows scan failed: %w", err)
		}

		tracklist.Tracks = append(tracklist.Tracks, &track)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows next failed: %w", err)
	}

	if tracklist.Name == "" {
		return nil, fmt.Errorf("could not find tracklist: %q", name)
	}

	return &tracklist, nil
}

func (s *Store) FindTracklistsByTrackIDCount(id string) (int, error) {
	var count struct {
		Count int
	}

	if err := s.DB.Get(&count, queries.FindTracklistsByTrackIDCount, id); err != nil {
		return -1, fmt.Errorf("db get failed: %w", err)
	}

	return count.Count, nil
}

func (s *Store) FindTracklistsByTrackID(id string, offset, limit int) ([]*Tracklist, error) {
	rows, err := s.Queryx(queries.FindTracklistsByTrackID, id, offset, limit)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}

	var tracklists []*Tracklist

	for rows.Next() {
		var tracklist Tracklist

		if err := rows.StructScan(&tracklist); err != nil {
			return nil, fmt.Errorf("rows scan failed: %w", err)
		}

		tracklists = append(tracklists, &tracklist)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows next failed: %w", err)
	}

	return tracklists, nil
}
