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

	switch os.Args[1] {
	case "create":
		if err := create(); err != nil {
			logger.Fatalf("error creating database: %v\n", err)
		}
	case "drop":
		if err := drop(); err != nil {
			logger.Fatalf("error dropping database: %v\n", err)
		}
	case "migrate":
		if err := migrate(logger); err != nil {
			logger.Fatalf("error applying migrations: %v\n", err)
		}
	case "rollback":
		if err := rollback(logger); err != nil {
			logger.Fatalf("error rolling back migrations: %v\n", err)
		}
	default:
		logger.Fatalf("error %q is not a valid command\n", os.Args[1])
	}
}
