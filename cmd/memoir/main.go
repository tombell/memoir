package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/charmbracelet/log"

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
  --version  Show the version number, then exit`

var (
	cfgpath = flag.String("config", ".env.dev.toml", "")
	version = flag.Bool("version", false, "")
)

func main() {
	logger := log.New(log.WithOutput(os.Stderr))

	flag.Usage = func() {
		logger.Print(helpText)
		os.Exit(2)
	}

	flag.Parse()

	if *version {
		logger.Print(fmt.Sprintf("memoir %s (%s)", Version, Commit))
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		flag.Usage()
	}

	switch os.Args[1] {
	case "run":
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
		logger.Fatal("invalid sub-command", "cmd", os.Args[1])
	}
}
