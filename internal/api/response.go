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

func (s *Server) addPaginationHeaders(w http.ResponseWriter, limit, page int32, total int64) {
	pages := int(math.Ceil(float64(total) / float64(limit)))

	w.Header().Add("Current-Page", strconv.FormatInt(int64(page), 10))
	w.Header().Add("Total-Pages", strconv.Itoa(pages))
}

func (s *Server) writeInternalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}

func (s *Server) writeBadRequest(w http.ResponseWriter, err error) {
	model := Error{Err: err.Error(), Msg: "Bad Request"}
	s.writeJSON(w, model, http.StatusBadRequest)
}

func (s *Server) writeUnprocessableContent(w http.ResponseWriter, err error) {
	model := Error{Err: err.Error(), Msg: "Unprocessable Content"}
	s.writeJSON(w, model, http.StatusUnprocessableEntity)
}

func (s *Server) writeNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (s *Server) writeJSON(w http.ResponseWriter, model any, status int) {
	resp, err := json.Marshal(model)
	if err != nil {
		s.writeInternalServerError(w)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	w.WriteHeader(status)
	w.Write(resp)
}
