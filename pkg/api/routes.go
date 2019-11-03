package api

func (s *Server) routes() {
	// Tracklists

	s.router.Handle("GET", "/tracklists",
		use(s.handleGetTracklists(), s.json, s.cors, s.instruments, s.requestID))

	s.router.Handle("GET", "/tracklists/:id",
		use(s.handleGetTracklist(), s.json, s.cors, s.instruments, s.requestID))

	// Tracks

	s.router.Handle("GET", "/tracks/:id/tracklists",
		use(s.handleGetTracklistsByTrackID(), s.json, s.cors, s.instruments, s.requestID))

	s.router.Handle("GET", "/tracks/mostplayed",
		use(s.handleGetMostPlayedTracks(), s.json, s.cors, s.instruments, s.requestID))

	s.router.Handle("GET", "/tracks/search",
		use(s.handleSearchTracks(), s.json, s.cors, s.instruments, s.requestID))
}
