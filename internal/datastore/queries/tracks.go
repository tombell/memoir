package queries

var (
	AddTrack                 string
	FindMostPlayedTracks     string
	FindTrackByArtistAndName string
	FindTrackByID            string
	FindTracksByQuery        string
)

func init() {
	AddTrack = query("sql/add_track.sql")
	FindMostPlayedTracks = query("sql/find_most_played_tracks.sql")
	FindTrackByArtistAndName = query("sql/find_track_by_artist_and_name.sql")
	FindTrackByID = query("sql/find_track_by_id.sql")
	FindTracksByQuery = query("sql/find_tracks_by_query.sql")
}
