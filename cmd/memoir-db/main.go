package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	// Import the Postgres driver for database/sql.
	_ "github.com/lib/pq"

	"github.com/tombell/memoir"
)

const (
	helpText = `usage: memoir-db <command> [<args>]

Commands:
  apply     apply migrations not currently applied to the database
  rollback  rolls back applied migrations from the database

Special options:
  --help     show this message, then exit
  --version  show the version number, then exit
`

	migrationsPath = `datastore/migrations`
)

var (
	version = flag.Bool("version", false, "")
)

func usage(text string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, text)
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
	case "apply":
		if err := apply(logger); err != nil {
			logger.Fatalf("error while applying migrations: %v\n", err)
		}
	case "rollback":
		if err := rollback(logger); err != nil {
			logger.Fatalf("error while rolling back migrations: %v\n", err)
		}
	default:
		logger.Fatalf("error: %q is not a valid command\n", os.Args[1])
	}
}
