package queries

var (
	AddTracklistTrack string
)

func init() {
	AddTracklistTrack = query("tracklist_tracks/add_tracklist_track.sql")
}
