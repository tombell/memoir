package api

func (s *Server) routes() {
	s.router.Handle("GET", "/tracklists", s.requestID(s.json(s.handleGetTracklists())))
	s.router.Handle("GET", "/tracklists/:id", s.requestID(s.json(s.handleGetTracklist())))
}
