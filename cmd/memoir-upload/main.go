package main

import (
	"flag"
	"fmt"
	"os"
)

const helpText = `usage: memoir-upload [args]

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
		fmt.Fprintf(os.Stdout, "memoir-upload %s (%s)\n", Version, Commit)
		os.Exit(0)
	}
}