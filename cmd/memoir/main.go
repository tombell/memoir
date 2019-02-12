package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tombell/memoir"
)

const helpText = `usage: memoir [args]

Special options:
  --help      show this message, then exit
  --version   show the version number, then exit
`

var (
	version = flag.Bool("version", false, "")
)

func usage() {
	fmt.Fprintf(os.Stderr, helpText)
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *version {
		fmt.Fprintf(os.Stdout, "memoir %s (%s)\n", memoir.Version, memoir.Commit)
		os.Exit(0)
	}
}
