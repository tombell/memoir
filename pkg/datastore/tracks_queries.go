package datastore

const (
	insertTrackSQL = `
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

	findTrackByIDSQL = `
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

	findTrackByArtistAndNameSQL = `
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

	findMostPlayedTracksSQL = `
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

	findTracksByQuerySQL = `
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

	updateTracksTSVectorSQL = `
		UPDATE tracks
		SET tsv =
		  setweight(to_tsvector(name), 'A') ||
		  setweight(to_tsvector(artist), 'B')`
)
