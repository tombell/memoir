package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
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

func exitIfError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", err)
		os.Exit(1)
	}
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

	records, err := readSeratoExport(args[0])
	exitIfError(err)

	t, err := time.Parse(dateTimeFormat, records[0][0])
	exitIfError(err)

	fmt.Fprintf(os.Stdout, "tracklist from %v\n\n", t.Format(dateTimeFormat))
}

func readSeratoExport(filepath string) ([][]string, error) {
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
