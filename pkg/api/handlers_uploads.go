package api

import (
	"net/http"
)

func (s *Server) handleUploadArtworkPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		file, header, err := r.FormFile("artwork")
		if err != nil {
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		uploaded, err := s.services.UploadArtwork(rid, file, header.Filename)
		if err != nil {
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.writeJSON(rid, w, uploaded)
	}
}
