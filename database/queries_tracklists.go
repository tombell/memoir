package database

const (
	sqlInsertTracklist = `
		INSERT INTO tracklists (
			id,
			date,
			created,
			updated
		) VALUES ($1, $2, $3, $4)`

	sqlInsertTracklistTrack = `
		INSERT INTO tracklist_tracks (
			id,
			tracklist_id,
			track_id
		) VALUES ($1, $2, $3)`

	sqlGetTracklistByID = `
		SELECT
			id,
			date,
			created,
			updated
		FROM tracklists
		WHERE id = $1
		LIMIT 1`
)
