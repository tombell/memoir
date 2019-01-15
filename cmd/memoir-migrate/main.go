package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/tombell/trek"
)

const helpText = `usage: memoir-migrate <command> [args]

Commands:
  apply       apply migrations not currently applied to the database
  rollback    rolls back applied migrations from the database

Special options:
  --help      show this message, then exit
  --version   show the version number, then exit
`

// TODO: update help text for apply
const applyHelpText = `usage: memoir-migrate apply [args]

Special options:
  --help      show this message, then exit
`

// TODO: update help text for rollback
const rollbackHelpText = `usage: memoir-migrate rollback [args]

Special options:
  --help      show this message, then exit
`

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
		fmt.Fprintf(os.Stdout, "memoir-migrate %s (%s)\n", Version, Commit)
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		flag.Usage()
	}

	logger := log.New(os.Stderr, "[memoir-migrate] ", log.LstdFlags)

	applyCmd := flag.NewFlagSet("apply", flag.ExitOnError)
	applyCmd.Usage = usage(applyHelpText)
	applyDsn := applyCmd.String("db", "", "")

	rollbackCmd := flag.NewFlagSet("rollback", flag.ExitOnError)
	rollbackCmd.Usage = usage(rollbackHelpText)
	rollbackDsn := rollbackCmd.String("db", "", "")

	switch os.Args[1] {
	case "apply":
		applyCmd.Parse(os.Args[2:])
	case "rollback":
		rollbackCmd.Parse(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "err: %q is not a valid command\n", os.Args[1])
		os.Exit(2)
	}

	if applyCmd.Parsed() {
		if err := trek.Apply(logger, "postgres", *applyDsn, "migrations"); err != nil {
			fmt.Fprintf(os.Stderr, "err: %v\n", err)
			os.Exit(1)
		}
	}

	if rollbackCmd.Parsed() {
		if err := trek.Rollback(logger, "postgres", *rollbackDsn, "migrations"); err != nil {
			fmt.Fprintf(os.Stderr, "err: %v\n", err)
			os.Exit(1)
		}
	}
}
