package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tombell/memoir"
	"github.com/tombell/memoir/datastore"
	"github.com/tombell/memoir/services"
)

const helpText = `usage: memoir-tracklist-delete [args]

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
		fmt.Fprintf(os.Stdout, "memoir-tracklist-delete %s (%s)\n", memoir.Version, memoir.Commit)
		os.Exit(0)
	}

	validateFlags()

	logger := log.New(os.Stderr, "", 0)

	store, err := datastore.New(*dsn)
	if err != nil {
		logger.Fatalf("err: %v\n", err)
	}
	defer store.Close()

	svc := services.Services{
		Logger:    logger,
		DataStore: store,
	}

	logger.Printf("deleting tracklist %s...\n", *tracklist)

	if err := svc.RemoveTracklist(*tracklist); err != nil {
		logger.Fatalf("error deleting tracklist: %v\n", err)
	}

	logger.Printf("deleted tracklist %s\n", *tracklist)
}
