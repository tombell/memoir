package api

func (s *Server) routes() {
	s.router.Handle("GET", "/tracklists", s.json(s.handleTracklistsIndex()))
}
