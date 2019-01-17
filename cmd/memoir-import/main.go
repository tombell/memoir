package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tombell/memoir/parser"
)

const helpText = `usage: memoir-import [args]

Special options:
  --help      show this message, then exit
  --version   show the version number, then exit
`

const (
	dateTimeFormat = "02/01/2006"
)

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
		fmt.Fprintf(os.Stdout, "memoir-import %s (%s)\n", Version, Commit)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
	}

	logger := log.New(os.Stderr, "[memoir-import] ", log.Ltime)

	records, err := parser.ParseSeratoExport(args[0])
	if err != nil {
		logger.Fatalf("err: %v\n", err)
	}

	t, err := time.Parse(dateTimeFormat, records[0][0])
	if err != nil {
		logger.Fatalf("err: %v\n", err)
	}

	logger.Printf("importing tracklist from %v...\n", t.Format(dateTimeFormat))
}
