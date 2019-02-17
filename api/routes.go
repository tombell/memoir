package api

func (s *Server) routes() {
	s.router.Handle("GET", "/tracklists", use(s.handleGetTracklists(), s.json, s.instruments, s.requestID))
	s.router.Handle("GET", "/tracklists/:id", use(s.handleGetTracklist(), s.json, s.instruments, s.requestID))
}
