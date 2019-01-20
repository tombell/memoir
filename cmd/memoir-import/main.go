package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/tombell/memoir/database"
	"github.com/tombell/memoir/services"
)

const helpText = `usage: memoir-import [args] <exported csv file>

  --name  name for the tracklist
  --db    connection string for connecting to the database

Special options:
  --help     show this message, then exit
  --version  show the version number, then exit
`

const (
	dateTimeFormat = "02/01/2006"
)

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

	records, err := parseSeratoExport(args[0])
	if err != nil {
		logger.Fatalf("err: %v\n", err)
	}

	t, err := time.Parse(dateTimeFormat, records[0][0])
	if err != nil {
		logger.Fatalf("err: %v\n", err)
	}

	logger.Printf("importing tracklist from %v as %q...\n", t.Format(dateTimeFormat), *name)

	db, err := database.Open(*dsn)
	if err != nil {
		logger.Fatalf("err: %v\n", err)
	}

	svc := services.Services{
		Logger: logger,
		DB:     db,
	}

	tracklist, err := svc.ImportTracklist(*name, t, records[1:])
	if err != nil {
		logger.Fatalf("err: %v\n", err)
	}

	logger.Println("finished importing")
	logger.Printf("created tracklist %q for %v with ID %q", tracklist.Name, tracklist.Date.Format(dateTimeFormat), tracklist.ID)
}

func parseSeratoExport(filepath string) ([][]string, error) {
	in, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(bytes.NewReader(in))

	var records [][]string

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	return records[1:], nil
}
