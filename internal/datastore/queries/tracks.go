package queries

const (
	AddTrack = `
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

	FindTrackByID = `
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
		WHERE id = $1
		LIMIT 1`

	FindTrackByArtistAndName = `
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

	FindMostPlayedTracks = `
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

	FindTracksByQuery = `
		SELECT
			id,
			artist,
			ts_headline(artist, q, 'StartSel=<<, StopSel=>>') as artist_highlighted,
			name,
			ts_headline(name, q, 'StartSel=<<, StopSel=>>') as name_highlighted,
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
				ts_rank(fts_name_and_artist, q) as rank,
				q
			FROM
				tracks,
				websearch_to_tsquery($1) q
			WHERE fts_name_and_artist @@ q
			ORDER BY rank DESC
		) as searched_tracks
		ORDER BY rank DESC
		LIMIT $2`
)