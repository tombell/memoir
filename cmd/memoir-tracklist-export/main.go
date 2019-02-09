package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tombell/memoir/database"
	"github.com/tombell/memoir/services"
)

const helpText = `usage: memoir-tracklist-export [args]

  --db         connection string for connecting to the database
  --tracklist  name for the tracklist being imported

Special options:
  --help      show this message, then exit
  --version   show the version number, then exit
`

const (
	dateTimeFormat = "02/01/2006"
)

var (
	dsn       = flag.String("db", "", "")
	tracklist = flag.String("tracklist", "", "")
	version   = flag.Bool("version", false, "")
)

func usage() {
	fmt.Fprintf(os.Stderr, helpText)
	os.Exit(2)
}

func validateFlags() {
	if *dsn == "" {
		flag.Usage()
	}

	if *tracklist == "" {
		flag.Usage()
	}
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *version {
		fmt.Fprintf(os.Stdout, "memoir-tracklist-export %s (%s)\n", Version, Commit)
		os.Exit(0)
	}

	validateFlags()

	logger := log.New(os.Stderr, "", 0)

	db, err := database.Open(*dsn)
	if err != nil {
		logger.Fatalf("err: %v\n", err)
	}
	defer db.Close()

	svc := services.Services{
		Logger: logger,
		DB:     db,
	}

	if err := svc.ExportTracklist(*tracklist, os.Stdout); err != nil {
		logger.Fatalf("error exporting tracklist: %v\n", err)
	}
}
