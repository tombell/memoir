package queries

var (
	AddTracklistTrack string
)

func init() {
	AddTracklistTrack = query("sql/add_tracklist_track.sql")
}
