package controllers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/tombell/memoir/internal/errors"
)

// Pagination parses page and per_page before narrowing SQL limits and offsets
// to int32. Missing values use page 1 and per_page 10.
func Pagination(pageParam, perPageParam string) (page, perPage int64, err error) {
	parse := func(field, value string, def int64) (int64, error) {
		if value == "" {
			return def, nil
		}
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil || n < 1 {
			return 0, errors.E("controllers[pagination]", errors.M{field: {"Must be a positive integer"}}, http.StatusBadRequest)
		}
		return n, nil
	}
	page, err = parse("page", pageParam, 1)
	if err != nil {
		return 0, 0, err
	}
	perPage, err = parse("per_page", perPageParam, 10)
	if err != nil {
		return 0, 0, err
	}
	if perPage > 100 {
		return 0, 0, errors.E("controllers[pagination]", errors.M{"per_page": {"Must be less than, or equal to 100"}}, http.StatusBadRequest)
	}
	if page-1 > math.MaxInt32/perPage {
		return 0, 0, errors.E("controllers[pagination]", errors.M{"page": {"Page offset is too large"}}, http.StatusBadRequest)
	}
	return page, perPage, nil
}
