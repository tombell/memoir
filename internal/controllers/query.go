package controllers

import (
	"strconv"
)

func IntQueryParam(param string, def int64) (int64, error) {
	if len(param) == 0 {
		return def, nil
	}

	page, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return 0, err
	}

	if page < def {
		return def, nil
	}

	return page, nil
}
