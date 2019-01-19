package database

const (
	sqlInsertTracklist = `
		INSERT INTO tracklists (
			id,
			name,
			date,
			created,
			updated
		) VALUES ($1, $2, $3, $4, $5)`

	sqlInsertTracklistTrack = `
		INSERT INTO tracklist_tracks (
			id,
			tracklist_id,
			track_id
		) VALUES ($1, $2, $3)`

	sqlGetTracklistByID = `
		SELECT
			id,
			name,
			date,
			created,
			updated
		FROM tracklists
		WHERE id = $1
		LIMIT 1`
)
