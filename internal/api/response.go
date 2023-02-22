package api

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
)

type Error struct {
	Err string `json:"err"`
	Msg string `json:"msg"`
}

func (s *Server) addPaginationHeaders(w http.ResponseWriter, limit, page, total int) {
	pages := int(math.Ceil(float64(total) / float64(limit)))

	w.Header().Add("Current-Page", strconv.Itoa(page))
	w.Header().Add("Total-Pages", strconv.Itoa(pages))
}

func (s *Server) writeInternalServerError(w http.ResponseWriter, err error) {
	s.Logger.Error("internal-server-error", "err", err)

	model := Error{Err: err.Error(), Msg: "internal service error"}

	s.writeJSON(w, model, http.StatusInternalServerError)
}

func (s *Server) writeBadRequest(w http.ResponseWriter, err error) {
	s.Logger.Error("bad-request", "err", err)

	model := Error{Err: err.Error(), Msg: "bad request"}

	s.writeJSON(w, model, http.StatusBadRequest)
}

func (s *Server) writeNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (s *Server) writeJSON(w http.ResponseWriter, model any, status int) {
	resp, err := json.Marshal(model)
	if err != nil {
		s.writeInternalServerError(w, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	w.WriteHeader(status)
	w.Write(resp)
}
