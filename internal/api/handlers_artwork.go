package api

import (
	"net/http"

	"github.com/tombell/mw"
)

func (s *Server) handlePostArtwork() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := mw.FindRequestID(r)

		file, header, err := r.FormFile("artwork")
		if err != nil {
			s.writeInternalServerError(rid, w, err)
			return
		}
		defer file.Close()

		uploaded, err := s.Services.UploadArtwork(rid, file, header.Filename)
		if err != nil {
			s.writeInternalServerError(rid, w, err)
			return
		}

		s.writeJSON(rid, w, uploaded)
	}
}
