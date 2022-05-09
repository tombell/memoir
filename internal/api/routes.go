package api

func (s *Server) routes() {
	api := []middleware{s.cors, s.logging, s.requestID}
	apiAuth := []middleware{s.cors, s.auth, s.logging, s.requestID}

	s.router.Handle("OPTIONS", "/...", use(s.handlePreflight(), s.cors))

	s.router.Handle("GET", "/tracklists", use(s.handleGetTracklists(), api...))
	s.router.Handle("POST", "/tracklists", use(s.handlePostTracklists(), apiAuth...))
	s.router.Handle("GET", "/tracklists/:id", use(s.handleGetTracklist(), api...))
	s.router.Handle("PATCH", "/tracklists/:id", use(s.handlePatchTracklist(), apiAuth...))

	s.router.Handle("GET", "/tracks/mostplayed", use(s.handleGetMostPlayedTracks(), api...))
	s.router.Handle("GET", "/tracks/search", use(s.handleSearchTracks(), api...))

	s.router.Handle("GET", "/tracks/:id/tracklists", use(s.handleGetTracklistsByTrack(), api...))
	s.router.Handle("GET", "/tracks/:id", use(s.handleGetTrack(), api...))

	s.router.Handle("POST", "/artwork", use(s.handlePostArtwork(), apiAuth...))
}
