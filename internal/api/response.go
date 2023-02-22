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

func (s *Server) writeInternalServerError(w http.ResponseWriter, err error) {
	s.Logger.Error("internal-server-error", "err", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (s *Server) writeNotFound(w http.ResponseWriter, r *http.Request) {
	s.Logger.Error("not-found", "path", r.URL.Path)
	w.WriteHeader(http.StatusNotFound)
}

func (s *Server) writeJSON(w http.ResponseWriter, model any) {
	resp, err := json.Marshal(model)
	if err != nil {
		s.writeInternalServerError(w, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
