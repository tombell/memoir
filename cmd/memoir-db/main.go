package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	// Import the Postgres driver for database/sql.
	_ "github.com/lib/pq"

	"github.com/tombell/trek"

	"github.com/tombell/memoir"
)

const (
	helpText = `usage: memoir-db <command> [args]

Commands:
  apply     apply migrations not currently applied to the database
  rollback  rolls back applied migrations from the database

Special options:
  --help     show this message, then exit
  --version  show the version number, then exit
`

	applyHelpText = `usage: memoir-db apply [args]

  --db  connection string for connecting to the database

Special options:
  --help  show this message, then exit
`

	rollbackHelpText = `usage: memoir-db rollback [args]

  --db  connection string for connecting to the database

Special options:
  --help  show this message, then exit
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
		cmd := flag.NewFlagSet("apply", flag.ExitOnError)
		cmd.Usage = usage(applyHelpText)
		dsn := cmd.String("db", "", "")

		if err := cmd.Parse(os.Args[2:]); err != nil {
			logger.Fatalf("error: %v\n", err)
		}

		if *dsn == "" {
			cmd.Usage()
		}

		if cmd.Parsed() {
			if err := trek.Apply(logger, "postgres", *dsn, migrationsPath); err != nil {
				logger.Fatalf("error migrating database: %v\n", err)
			}
		}
	case "rollback":
		cmd := flag.NewFlagSet("rollback", flag.ExitOnError)
		cmd.Usage = usage(rollbackHelpText)
		dsn := cmd.String("db", "", "")

		if err := cmd.Parse(os.Args[2:]); err != nil {
			logger.Fatalf("error: %v\n", err)
		}

		if *dsn == "" {
			cmd.Usage()
		}

		if cmd.Parsed() {
			if err := trek.Rollback(logger, "postgres", *dsn, migrationsPath); err != nil {
				logger.Fatalf("error rolling back database: %v\n", err)
			}
		}
	default:
		logger.Fatalf("error: %q is not a valid command\n", os.Args[1])
	}
}
