package main

import (
	"flag"
	"log"
	"os"

	"github.com/tombell/trek"
)

const (
	applyHelpText = `usage: memoir-db apply [<args>]

  --db  connection string for connecting to the database

Special options:
  --help  show this message, then exit
`
	rollbackHelpText = `usage: memoir-db rollback [<args>]

  --db  connection string for connecting to the database

Special options:
  --help  show this message, then exit
`
)

func apply(logger *log.Logger) error {
	cmd := flag.NewFlagSet("apply", flag.ExitOnError)
	cmd.Usage = usage(applyHelpText)
	dsn := cmd.String("db", "", "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return err
	}

	if *dsn == "" {
		cmd.Usage()
	}

	if cmd.Parsed() {
		return trek.Apply(logger, "postgres", *dsn, migrationsPath)
	}

	return nil
}

func rollback(logger *log.Logger) error {
	cmd := flag.NewFlagSet("rollback", flag.ExitOnError)
	cmd.Usage = usage(rollbackHelpText)
	dsn := cmd.String("db", "", "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return err
	}

	if *dsn == "" {
		cmd.Usage()
	}

	if cmd.Parsed() {
		return trek.Rollback(logger, "postgres", *dsn, migrationsPath)
	}

	return nil
}
