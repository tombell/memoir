package api

import (
	"net/http"
)

func (s *Server) handlePostArtwork() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("artwork")
		if err != nil {
			s.writeInternalServerError(w)
			return
		}
		defer file.Close()

		uploaded, exists, err := s.services.UploadArtwork(file, header.Filename)
		if err != nil {
			s.writeInternalServerError(w)
			return
		}

		status := http.StatusCreated
		if exists {
			status = http.StatusOK
		}

		s.writeJSON(w, uploaded, status)
	}
}
