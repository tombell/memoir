package queries

const (
	// InsertTracklistTrack ...
	InsertTracklistTrack = `
		INSERT INTO tracklist_tracks (
			id,
			tracklist_id,
			track_id,
			track_number
		) VALUES ($1, $2, $3, $4)`
)
