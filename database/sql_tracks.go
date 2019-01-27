package database

const (
	sqlInsertTrack = `
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

	sqlGetTrackByArtistAndName = `
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

	sqlGetTracksByGenre = `
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
		WHERE genre = $1`
)
