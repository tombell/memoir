package api

func (s *Server) routes() {
	s.router.Handle("GET", "/tracklists", s.json(s.handleGetTracklists()))
	s.router.Handle("GET", "/tracklists/:id", s.json(s.handleGetTracklist()))
}
