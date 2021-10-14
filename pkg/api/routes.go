package api

func (s *Server) routes() {
	// CORS

	s.router.Handle("OPTIONS", "/...",
		use(s.handlePreflight(), s.cors))

	// Health

	s.router.Handle("GET", "/healthz",
		use(s.handleHealth(), s.logging, s.requestID))

	// Tracklists

	s.router.Handle("GET", "/tracklists",
		use(s.handleTracklistsGet(), s.json, s.cors, s.logging, s.requestID))

	s.router.Handle("POST", "/tracklists",
		use(s.handleTracklistsPost(), s.json, s.cors, s.auth, s.logging, s.requestID))

	s.router.Handle("GET", "/tracklists/:id",
		use(s.handleTracklistGet(), s.json, s.cors, s.logging, s.requestID))

	s.router.Handle("PATCH", "/tracklists/:id",
		use(s.handleTracklistPatch(), s.json, s.cors, s.auth, s.logging, s.requestID))

	// Tracks

	s.router.Handle("GET", "/tracks/:id",
		use(s.handleTrackGet(), s.json, s.cors, s.logging, s.requestID))

	s.router.Handle("GET", "/tracks/:id/tracklists",
		use(s.handleTracklistsByTrackGet(), s.json, s.cors, s.logging, s.requestID))

	s.router.Handle("GET", "/tracks/mostplayed",
		use(s.handleTracksMostPlayedGet(), s.json, s.cors, s.logging, s.requestID))

	s.router.Handle("GET", "/tracks/search",
		use(s.handleTracksSearchGet(), s.json, s.cors, s.logging, s.requestID))

	// Uploads

	s.router.Handle("POST", "/uploads/artwork",
		use(s.handleUploadArtworkPost(), s.json, s.cors, s.auth, s.logging, s.requestID))
}
