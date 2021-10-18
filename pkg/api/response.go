package api

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
)

func (s *Server) addPaginationHeaders(w http.ResponseWriter, per, page, total int) {
	pages := int(math.Ceil(float64(total) / float64(per)))

	w.Header().Add("Current-Page", strconv.Itoa(page))
	w.Header().Add("Total-Pages", strconv.Itoa(pages))
}

func (s *Server) writeError(rid string, w http.ResponseWriter, err error) {
	s.services.Logger.Printf("[%s] error=%s\n", rid, err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (s *Server) writeJSON(rid string, w http.ResponseWriter, model interface{}) {
	resp, err := json.Marshal(model)
	if err != nil {
		s.writeError(rid, w, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(resp)
}
