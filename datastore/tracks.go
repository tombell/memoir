package datastore

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	sqlAddTrack = `
		INSERT INTO tracks (
			id,
			artist,
			name,
			genre,
			bpm,
			key,
			created,
			updated
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	sqlGetTrackByID = `
		SELECT
			id,
			artist,
			name,
			genre,
			bpm,
			key
			created,
			updated
		FROM tracks
		WHERE id = $1
		LIMIT 1`

	sqlFindTrackByArtistAndName = `
		SELECT
			id,
			artist,
			name,
			genre,
			bpm,
			key,
			created,
			updated
		FROM tracks
		WHERE artist = $1
		AND name = $2
		LIMIT 1`

	sqlFindMostPlayedTracks = `
		SELECT
			t.id,
			t.artist,
			t.name,
			t.genre,
			t.bpm,
			t.key,
			t.created,
			t.updated,
			count(t.id) as played
		FROM tracks t
		JOIN tracklist_tracks tt ON tt.track_id = t.id
		GROUP BY t.id
		ORDER BY played DESC
		LIMIT $1`

	sqlFindTracksByQuery = `
		SELECT
			id,
			artist,
			ts_headline(artist, q) as artist_highlighted,
			name,
			ts_headline(name, q) as name_highlighted,
			genre,
			bpm,
			key,
			created,
			updated
		FROM (
			SELECT
				id,
				artist,
				name,
				genre,
				bpm,
				key,
				created,
				updated,
				ts_rank(tsv, q) as rank,
				q
			FROM
				tracks,
				plainto_tsquery($1) q
			WHERE tsv @@ q
			ORDER BY rank DESC
		) as searched_tracks
		ORDER BY rank DESC`
)

// Track contains data about a track row in the database.
type Track struct {
	ID      string
	Artist  string
	Name    string
	Genre   string
	BPM     float64
	Key     string
	Created time.Time
	Updated time.Time

	Played int
}

// TrackSearchResult contains data about track search result matching on the
// artist or name.
type TrackSearchResult struct {
	Track
	ArtistHighlighted string
	NameHighlighted   string
}

// AddTrack adds a new track into the database.
func (s *Store) AddTrack(tx *sql.Tx, track *Track) error {
	_, err := tx.Exec(sqlAddTrack,
		track.ID,
		track.Artist,
		track.Name,
		track.Genre,
		track.BPM,
		track.Key,
		track.Created,
		track.Updated)

	if err != nil {
		return fmt.Errorf("tx exec failed: %w", err)
	}

	return nil
}

// GetTrack gets a track with the given ID from the database.
func (s *Store) GetTrack(id string) (*Track, error) {
	var track Track

	err := s.QueryRowx(sqlGetTrackByID, id).StructScan(&track)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("query row failed: %w", err)
	default:
		return &track, nil
	}
}

// FindTrackByArtistAndName finds a track with the given artist and name in the
// database.
func (s *Store) FindTrackByArtistAndName(artist, name string) (*Track, error) {
	var track Track

	err := s.QueryRowx(sqlFindTrackByArtistAndName, artist, name).StructScan(&track)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("query row failed: %w", err)
	default:
		return &track, nil
	}
}

// FindMostPlayedTracks finds the tracks that are most played, limiting it to
// the given count in the database.
func (s *Store) FindMostPlayedTracks(limit int) ([]*Track, error) {
	rows, err := s.Queryx(sqlFindMostPlayedTracks, limit)
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

	var tracks []*Track

	for rows.Next() {
		var track Track

		if err := rows.StructScan(&track); err != nil {
			return nil, fmt.Errorf("rows struct scan failed: %w", err)
		}

		tracks = append(tracks, &track)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows next failed: %w", err)
	}

	return tracks, nil
}

// FindTracksByQuery finds the tracks that have artists or names matching the
// given query in the database.
func (s *Store) FindTracksByQuery(query string) ([]*TrackSearchResult, error) {
	rows, err := s.Queryx(sqlFindTracksByQuery, query)
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

	var tracks []*TrackSearchResult

	for rows.Next() {
		var track TrackSearchResult

		if err := rows.StructScan(&track); err != nil {
			return nil, fmt.Errorf("rows struct scan failed: %w", err)
		}

		tracks = append(tracks, &track)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows next failed: %w", err)
	}

	return tracks, nil
}
