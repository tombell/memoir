package database

const (
	sqlInsertTrack = `
		INSERT INTO tracks (
			id,
			artist,
			name,
			genre,
			bpm,
			key
		) VALUES ($1, $2, $3, $4, $5, $6);`

	sqlGetTrackByID = `
		SELECT
			id,
			artist,
			name,
			genre,
			bpm,
			key
		FROM tracks
		WHERE id = $1
		LIMIT 1`

	sqlGetTracksByGenre = `
		SELECT
			id,
			artist,
			name,
			genre,
			bpm,
			key
		FROM tracks
		WHERE genre = $1`
)
