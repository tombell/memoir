package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) writeJSON(rid string, w http.ResponseWriter, model interface{}) {
	resp, err := json.Marshal(model)
	if err != nil {
		s.services.Logger.Printf("[%s] error=%s\n", rid, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(resp)
}
