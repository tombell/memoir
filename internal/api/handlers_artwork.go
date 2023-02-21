package api

import (
	"net/http"
)

func (s *Server) handlePostArtwork() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("artwork")
		if err != nil {
			s.writeInternalServerError(w, err)
			return
		}
		defer file.Close()

		uploaded, err := s.UploadArtwork(file, header.Filename)
		if err != nil {
			s.writeInternalServerError(w, err)
			return
		}

		s.writeJSON(w, uploaded)
	}
}
