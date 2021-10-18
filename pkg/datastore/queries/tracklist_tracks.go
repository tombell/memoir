package queries

const (
	// AddTracklistTrack adds the given attributes as a new tracklist to track
	// mapping.
	AddTracklistTrack = `
		INSERT INTO tracklist_tracks (
			id,
			tracklist_id,
			track_id,
			track_number
		) VALUES ($1, $2, $3, $4)`
)
