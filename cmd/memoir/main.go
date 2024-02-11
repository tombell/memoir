package main

import (
	"flag"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/tombell/memoir/cmd/memoir/commands"
)

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		flag.Usage()
		return
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	switch os.Args[1] {
	case "run", "r":
		commands.Run(logger)
	case "db:create":
		commands.DatabaseCreate(logger)
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
