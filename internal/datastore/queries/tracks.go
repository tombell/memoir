package queries

var (
	AddTrack                 string
	FindMostPlayedTracks     string
	FindTrackByArtistAndName string
	FindTrackByID            string
	FindTracksByQuery        string
)

func init() {
	AddTrack = query("tracks/add_track.sql")
	FindMostPlayedTracks = query("tracks/find_most_played_tracks.sql")
	FindTrackByArtistAndName = query("tracks/find_track_by_artist_and_name.sql")
	FindTrackByID = query("tracks/find_track_by_id.sql")
	FindTracksByQuery = query("tracks/find_tracks_by_query.sql")
}
