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

Special options:
  --help     show this message, then exit
  --version  show the version number, then exit
`

var (
	name    = flag.String("name", "", "")
	dsn     = flag.String("db", "", "")
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
		fmt.Fprintf(os.Stdout, "memoir-delete %s (%s)\n", Version, Commit)
		os.Exit(0)
	}

	if *dsn == "" || *name == "" {
		flag.Usage()
	}

	logger := log.New(os.Stderr, "[memoir-delete] ", log.Ltime)

	db, err := database.Open(*dsn)
	if err != nil {
		logger.Fatalf("err: %v\n", err)
	}

	svc := services.Services{
		Logger: logger,
		DB:     db,
	}

	if err := svc.RemoveTracklist(*name); err != nil {
		logger.Fatalf("err: %v\n", err)
	}

	logger.Printf("finished deleting tracklist %q\n", *name)
}