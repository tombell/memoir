package datastore

const (
	insertMixUploadSQL = `
		INSERT INTO mix_uploads (
			id,
			tracklist_id,
			filename,
			location,
			created,
			updated
		) VALUES ($1, $2, $3, $4, $5, $6)`

	deleteMixUploadSQL = `
		DELETE FROM mix_uploads
		WHERE id = $1`
)
