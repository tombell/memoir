package database

const (
	sqlInsertMixUpload = `
		INSERT INTO mix_uploads (
			id,
			tracklist_id,
			filename,
			location,
			created,
			updated
		) VALUES ($1, $2, $3, $4, $5, $6)`
)
