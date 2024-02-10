package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/tombell/memoir/cmd/memoir/commands"
)

const helpText = `usage: memoir <command> [<args>]

Commands:
  run          Run the API server
  db:create    Create the database
  db:drop      Drop the database
  db:migrate   Migrate the database
  db:rollback  Rolls back applied migrations from the database

Special options:
  --help     Show this message, then exit
  --version  Show the version number, then exit
`

var (
	cfgpath = flag.String("config", ".env.dev.toml", "")
	version = flag.Bool("version", false, "")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, helpText)
		os.Exit(2)
	}

	flag.Parse()

	if *version {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("memoir %s (%s)\n", Version, Commit))
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		flag.Usage()
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	switch os.Args[1] {
	case "run", "r":
		commands.RunCommand(logger)
	case "db:create":
		commands.DatabaseCreateCommand(logger)
	case "db:drop":
		commands.DatabaseDrop(logger)
	case "db:migrate":
		commands.DatabaseMigrate(logger)
	case "db:rollback":
		commands.DatabaseRollback(logger)
	default:
		logger.Error("invalid sub-command", "cmd", os.Args[1])
	}
}
