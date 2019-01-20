package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tombell/memoir/database"
	"github.com/tombell/memoir/services"
)

const helpText = `usage: memoir-delete [args]

  --db         connection string for connecting to the database
  --tracklist  name of the tracklist to delete

Special options:
  --help     show this message, then exit
  --version  show the version number, then exit
`

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
		fmt.Fprintf(os.Stdout, "memoir-delete %s (%s)\n", Version, Commit)
		os.Exit(0)
	}

	validateFlags()

	logger := log.New(os.Stderr, "", 0)

	db, err := database.Open(*dsn)
	if err != nil {
		logger.Fatalf("err: %v\n", err)
	}

	svc := services.Services{
		Logger: logger,
		DB:     db,
	}

	if err := svc.RemoveTracklist(*tracklist); err != nil {
		logger.Fatalf("err: %v\n", err)
	}

	logger.Printf("finished deleting tracklist %q\n", *tracklist)
}
