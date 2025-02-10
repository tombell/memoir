package controllers

import (
	"strconv"
)

// ParamAsInt is a helper function to convert a string param to an int64, and if
// it's empty or less than or equal to zero, return the default value.
func ParamAsInt(param string, def int64) (int64, error) {
	if len(param) == 0 {
		return def, nil
	}

	num, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return 0, err
	}

	if num <= 0 {
		return def, nil
	}

	return num, nil
}
