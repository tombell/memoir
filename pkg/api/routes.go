package api

func (s *Server) routes() {
	s.router.Handle("OPTIONS", "/...", use(s.handlePreflight(), s.cors))
	s.router.Handle("GET", "/healthz", use(s.handleHealth(), s.instruments, s.requestID))

	// Tracklists

	s.router.Handle("GET", "/tracklists",
		use(s.handleGetTracklists(), s.json, s.cors, s.instruments, s.requestID))

	s.router.Handle("POST", "/tracklists",
		use(s.handleImportTracklist(), s.json, s.cors, s.auth, s.instruments, s.requestID))

	s.router.Handle("GET", "/tracklists/:id",
		use(s.handleGetTracklist(), s.json, s.cors, s.instruments, s.requestID))

	s.router.Handle("PATCH", "/tracklists/:id",
		use(s.handleUpdateTracklist(), s.json, s.cors, s.auth, s.instruments, s.requestID))

	// Tracks

	s.router.Handle("GET", "/tracks/:id/tracklists",
		use(s.handleGetTracklistsByTrackID(), s.json, s.cors, s.instruments, s.requestID))

	s.router.Handle("GET", "/tracks/mostplayed",
		use(s.handleGetMostPlayedTracks(), s.json, s.cors, s.instruments, s.requestID))

	s.router.Handle("GET", "/tracks/search",
		use(s.handleSearchTracks(), s.json, s.cors, s.instruments, s.requestID))

	// Uploads

	s.router.Handle("POST", "/uploads/artwork",
		use(s.handleUploadArtwork(), s.json, s.cors, s.auth, s.instruments, s.requestID))

}
