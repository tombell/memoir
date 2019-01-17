package database

const (
	sqlInsertTracklist = `
		INSERT INTO tracklists (
			id,
			created,
			updated
		) VALUES ($1, $2, $3)`

	sqlGetTracklistByID = `
		SELECT
			id,
			created,
			updated
		FROM tracklists
		WHERE id = $1
		LIMIT 1`
)
