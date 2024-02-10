package commands

import (
	"regexp"
)

var matchDBNameRegexp = regexp.MustCompile(`dbname=([a-zA-Z0-9_]+)`)
