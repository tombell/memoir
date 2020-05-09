package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	// Import the Postgres driver for database/sql.
	_ "github.com/lib/pq"

	"github.com/tombell/memoir"
)

const (
	helpText = `usage: memoir-db <command> [<args>]

Commands:
  create     Create the database
  drop       Drop the database
  migrate    Migrate the database
  rollback   Rolls back applied migrations from the database

Special options:
  --help     Show this message, then exit
  --version  Show the version number, then exit
`
)

var (
	matchDBNameRegexp = regexp.MustCompile(`dbname=([a-zA-Z0-9_]+)`)

	version = flag.Bool("version", false, "")
)

func usage(text string) func() {
	return func() {
		fmt.Fprint(os.Stderr, text)
		os.Exit(2)
	}
}

func main() {
	flag.Usage = usage(helpText)
	flag.Parse()

	if *version {
		fmt.Fprintf(os.Stdout, "memoir-db %s (%s)\n", memoir.Version, memoir.Commit)
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		flag.Usage()
	}

	logger := log.New(os.Stderr, "", 0)

	var err error

	switch os.Args[1] {
	case "create":
		err = create()
	case "drop":
		err = drop()
	case "migrate":
		err = migrate()
	case "rollback":
		err = rollback()
	default:
		err = fmt.Errorf("%q is not a valid command", os.Args[1])
	}

	if err != nil {
		logger.Fatalf("error: %v", err)
	}
}
