package api

import (
	"net/http"
	"strconv"
)

func pageQueryParam(r *http.Request) (int, error) {
	params := r.URL.Query()

	page := params.Get("page")
	if page == "" {
		page = "1"
	}

	return strconv.Atoi(page)
}
