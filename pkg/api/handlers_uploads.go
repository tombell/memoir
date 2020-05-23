package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleUploadArtwork() http.HandlerFunc {
	type response struct {
		Key string `json:"artwork"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		rid := getRequestID(r)

		file, header, err := r.FormFile("artwork")
		if err != nil {
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		key, err := s.services.UploadArtwork(rid, file, header.Filename)
		if err != nil {
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(&response{Key: key})
		if err != nil {
			s.services.Logger.Printf("[%s] error=%s\n", rid, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(resp)
	}
}
