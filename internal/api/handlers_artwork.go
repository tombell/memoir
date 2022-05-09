package api

import (
	"net/http"
)

func (s *Server) handlePostArtwork() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		file, header, err := r.FormFile("artwork")
		if err != nil {
			s.writeInternalServerError(rid, w, err)
			return
		}
		defer file.Close()

		uploaded, err := s.services.UploadArtwork(rid, file, header.Filename)
		if err != nil {
			s.writeInternalServerError(rid, w, err)
			return
		}

		s.writeJSON(rid, w, uploaded)
	}
}
