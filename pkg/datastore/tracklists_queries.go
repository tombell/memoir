package datastore

const (
	insertTracklistSQL = `
		INSERT INTO tracklists (
			id,
			name,
			url,
			artwork,
			date,
			created,
			updated
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	deleteTracklistSQL = `
		DELETE FROM tracklists
		WHERE id = $1`

	addArtworkToTracklistSQL = `
		UPDATE tracklists
		SET artwork = $1
		WHERE id = $2`

	getTracklistsSQL = `
		SELECT
			tl.*,
			count(tl.id) as track_count
		FROM tracklists tl
		JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
		GROUP BY tl.id
		ORDER BY tl.date DESC`

	findTracklistByIDSQL = `
		SELECT
			id,
			name,
			date,
			created,
			updated
		FROM tracklists
		WHERE id = $1
		LIMIT 1`

	findTracklistWithTracksByIDSQL = `
		SELECT
			tl.*,
			t.id as track_id,
			t.artist,
			t.name,
			t.genre,
			t.bpm,
			t.key,
			t.created,
			t.updated
		FROM tracklists tl
		JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
		JOIN tracks t ON t.id = tt.track_id
		WHERE tl.id = $1
		ORDER BY tt.track_number ASC`

	findTracklistByNameSQL = `
		SELECT
			id,
			name,
			date,
			created,
			updated
		FROM tracklists
		WHERE name = $1
		LIMIT 1`

	findTracklistWithTracksByNameSQL = `
		SELECT
			tl.*,
			t.id as track_id,
			t.artist,
			t.name,
			t.genre,
			t.bpm,
			t.key,
			t.created,
			t.updated
		FROM tracklists tl
		JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
		JOIN tracks t ON t.id = tt.track_id
		WHERE tl.name = $1
		ORDER BY tt.track_number ASC`

	findTracklistByTrackIDSQL = `
		SELECT tl.*, (
			SELECT count(id)
			FROM tracklist_tracks
			WHERE tracklist_tracks.tracklist_id = tl.id
		) as track_count
		FROM tracklists tl
		JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
		WHERE tt.track_id = $1
		ORDER BY tl.date DESC`
)
