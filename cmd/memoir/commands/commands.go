package commands

import (
	"fmt"
	"os"
	"regexp"
)

var matchDBNameRegexp = regexp.MustCompile(`dbname=([a-zA-Z0-9_]+)`)

func usageText(text string) func() {
	return func() {
		fmt.Fprint(os.Stderr, text)
		os.Exit(2)
	}
}
