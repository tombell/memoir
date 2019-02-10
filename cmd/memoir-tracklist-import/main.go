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

	"github.com/tombell/memoir/datastore"
	"github.com/tombell/memoir/services"
)

const helpText = `usage: memoir-tracklist-import [args] <exported csv file>

  --db         connection string for connecting to the database
  --tracklist  name for the tracklist being imported

Special options:
  --help     show this message, then exit
  --version  show the version number, then exit
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
		fmt.Fprintf(os.Stdout, "memoir-tracklist-import %s (%s)\n", Version, Commit)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
	}

	validateFlags()

	logger := log.New(os.Stderr, "", 0)

	logger.Printf("importing tracklist %s...\n", *tracklist)

	records, err := parseSeratoExport(args[0])
	if err != nil {
		logger.Fatalf("error importing tracklist: %v\n", err)
	}

	date, err := time.Parse(dateTimeFormat, records[0][0])
	if err != nil {
		logger.Fatalf("error importing tracklist: %v\n", err)
	}

	store, err := datastore.New(*dsn)
	if err != nil {
		logger.Fatalf("error importing tracklist: %v\n", err)
	}
	defer store.Close()

	svc := services.Services{
		Logger:    logger,
		DataStore: store,
	}

	tracklist, err := svc.ImportTracklist(*tracklist, date, records[1:])
	if err != nil {
		logger.Fatalf("error importing tracklist: %v\n", err)
	}

	logger.Printf("imported tracklist %s\n", tracklist.Name)
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
