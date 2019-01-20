package main

import (
	"flag"
	"fmt"
	"os"
)

const helpText = `usage: memoir-upload [args] <path to mix mp3>

  --db    connection string for connecting to the database
  --name  name of the tracklist to associate the uploaded mix with

Special options:
  --help      show this message, then exit
  --version   show the version number, then exit
`

var (
	dsn     = flag.String("db", "", "")
	name    = flag.String("name", "", "")
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

	if *dsn == "" {
		flag.Usage()
	}
}
