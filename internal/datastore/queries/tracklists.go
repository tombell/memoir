package queries

var (
	AddTracklist                  string
	UpdateTracklist               string
	GetTracklists                 string
	GetTracklistsCount            string
	FindTracklistByID             string
	FindTracklistByName           string
	FindTracklistWithTracksByID   string
	FindTracklistWithTracksByName string
	FindTracklistsByTrackID       string
	FindTracklistsByTrackIDCount  string
)

func init() {
	AddTracklist = query("tracklists/add_tracklist.sql")
	UpdateTracklist = query("tracklists/update_tracklist.sql")
	GetTracklists = query("tracklists/get_tracklists.sql")
	GetTracklistsCount = query("tracklists/get_tracklists_count.sql")
	FindTracklistByID = query("tracklists/find_tracklist_by_id.sql")
	FindTracklistByName = query("tracklists/find_tracklist_by_name.sql")
	FindTracklistWithTracksByID = query("tracklists/find_tracklist_with_tracks_by_id.sql")
	FindTracklistWithTracksByName = query("tracklists/find_tracklist_with_tracks_by_name.sql")
	FindTracklistsByTrackID = query("tracklists/find_tracklists_by_track_id.sql")
	FindTracklistsByTrackIDCount = query("tracklists/find_tracklists_by_track_id_count.sql")
}
