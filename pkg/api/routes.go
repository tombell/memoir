package api

func (s *Server) routes() {
	s.router.Handle("GET", "/tracklists", use(s.handleGetTracklists(), s.cors, s.json, s.instruments, s.requestID))
	s.router.Handle("GET", "/tracklists/:id", use(s.handleGetTracklist(), s.cors, s.json, s.instruments, s.requestID))

	s.router.Handle("GET", "/tracks/:id/tracklists", use(s.handleGetTracklistsByTrack(), s.cors, s.json, s.instruments, s.requestID))
	s.router.Handle("GET", "/tracks/mostplayed", use(s.handleGetMostPlayedTracks(), s.cors, s.json, s.instruments, s.requestID))
	s.router.Handle("GET", "/tracks/search", use(s.handleSearchTracks(), s.cors, s.json, s.instruments, s.requestID))
}