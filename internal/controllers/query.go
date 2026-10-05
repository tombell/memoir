package controllers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/tombell/memoir/internal/errors"
)

const (
	DefaultPerPage int64 = 10
	MaxPerPage     int64 = 100
)

// ParamAsInt converts a positive integer parameter, using def if it is absent.
func ParamAsInt(param string, def int64) (int64, error) {
	if len(param) == 0 {
		return def, nil
	}

	num, err := strconv.ParseInt(param, 10, 64)
	if err != nil || num <= 0 {
		return 0, errors.E(errors.Op("controllers[query]"), http.StatusBadRequest,
			errors.M{"message": {"Query parameter must be a positive integer"}})
	}

	return num, nil
}

// PaginationParams defaults omitted values to page 1 and per_page 10. It bounds
// per_page to 100 and the resulting offset to int32 before stores calculate it.
// Pointers distinguish omitted parameters from explicitly empty values.
func PaginationParams(pageParam, perPageParam *string) (page, perPage int64, err error) {
	page, err = paginationParam("page", pageParam, 1)
	if err != nil {
		return 0, 0, err
	}

	perPage, err = paginationParam("per_page", perPageParam, DefaultPerPage)
	if err != nil {
		return 0, 0, err
	}
	if perPage > MaxPerPage {
		return 0, 0, queryError("per_page", "Must be less than, or equal to 100")
	}

	// Divide before comparing so even an int64-sized page cannot overflow.
	if page-1 > math.MaxInt32/perPage {
		return 0, 0, queryError("page", "Page offset must be less than, or equal to 2147483647")
	}

	return page, perPage, nil
}

func paginationParam(name string, param *string, def int64) (int64, error) {
	if param == nil {
		return def, nil
	}
	if *param == "" {
		return 0, queryError(name, "Must be a positive integer")
	}
	num, err := ParamAsInt(*param, def)
	if err != nil {
		return 0, queryError(name, "Must be a positive integer")
	}
	return num, nil
}

func queryError(name, message string) error {
	return errors.E(errors.Op("controllers[query]"), http.StatusBadRequest,
		errors.M{name: {message}})
}
