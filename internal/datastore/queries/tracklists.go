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
	AddTracklist = query("sql/add_tracklist.sql")
	UpdateTracklist = query("sql/update_tracklist.sql")
	GetTracklists = query("sql/get_tracklists.sql")
	GetTracklistsCount = query("sql/get_tracklists_count.sql")
	FindTracklistByID = query("sql/find_tracklist_by_id.sql")
	FindTracklistByName = query("sql/find_tracklist_by_name.sql")
	FindTracklistWithTracksByID = query("sql/find_tracklist_with_tracks_by_id.sql")
	FindTracklistWithTracksByName = query("sql/find_tracklist_with_tracks_by_name.sql")
	FindTracklistsByTrackID = query("sql/find_tracklists_by_track_id.sql")
	FindTracklistsByTrackIDCount = query("sql/find_tracklists_by_track_id_count.sql")
}
