package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tombell/memoir/database"
	"github.com/tombell/memoir/parser"
	"github.com/tombell/memoir/services"
)

const helpText = `usage: memoir-import [args] <exported csv file>

TODO: flags

Special options:
  --help      show this message, then exit
  --version   show the version number, then exit
`

const (
	dateTimeFormat = "02/01/2006"
)

var (
	version = flag.Bool("version", false, "")
	dsn     = flag.String("db", "", "")
	name    = flag.String("name", "", "")
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

	if *dsn == "" || *name == "" {
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

	db, err := database.Open(*dsn)
	if err != nil {
		logger.Fatalf("err: %v\n", err)
	}

	svc := services.Services{
		Logger: logger,
		DB:     db,
	}

	if err := svc.ImportTracklist(*name, t, records[2:]); err != nil {
		logger.Fatalf("err: %v\n", err)
	}

	logger.Println("finished importing")
}
