package api

func (s *Server) routes() {
	api := []middleware{s.cors, s.logging, s.requestID}
	apiAuth := []middleware{s.cors, s.auth, s.logging, s.requestID}

	// CORS

	s.router.Handle("OPTIONS", "/...", use(s.handlePreflight(), s.cors))

	// Health

	s.router.Handle("GET", "/healthz", use(s.handleHealth(), s.logging, s.requestID))

	// Tracklists

	s.router.Handle("GET", "/tracklists", use(s.handleGetTracklists(), api...))
	s.router.Handle("POST", "/tracklists", use(s.handlePostTracklists(), apiAuth...))
	s.router.Handle("GET", "/tracklists/:id", use(s.handleGetTracklist(), api...))
	s.router.Handle("PATCH", "/tracklists/:id", use(s.handlePatchTracklist(), apiAuth...))

	// Tracks

	s.router.Handle("GET", "/tracks/mostplayed", use(s.handleGetMostPlayedTracks(), api...))
	s.router.Handle("GET", "/tracks/search", use(s.handleSearchTracks(), api...))

	s.router.Handle("GET", "/tracks/:id/tracklists", use(s.handleGetTracklistsByTrack(), api...))
	s.router.Handle("GET", "/tracks/:id", use(s.handleGetTrack(), api...))

	// Uploads

	s.router.Handle("POST", "/uploads/artwork", use(s.handlePostArtwork(), apiAuth...))
}
