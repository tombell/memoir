package api

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
)

func (s *Server) addPaginationHeaders(w http.ResponseWriter, limit, page, total int) {
	pages := int(math.Ceil(float64(total) / float64(limit)))

	w.Header().Add("Current-Page", strconv.Itoa(page))
	w.Header().Add("Total-Pages", strconv.Itoa(pages))
}

func (s *Server) writeInternalServerError(rid string, w http.ResponseWriter, err error) {
	s.services.Logger.Printf("[%s] error=%s", rid, err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (s *Server) writeNotFound(rid string, w http.ResponseWriter, msg string) {
	s.services.Logger.Printf("[%s] could not find %s", rid, msg)
	w.WriteHeader(http.StatusNotFound)
}

func (s *Server) writeJSON(rid string, w http.ResponseWriter, model interface{}) {
	resp, err := json.Marshal(model)
	if err != nil {
		s.writeInternalServerError(rid, w, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
