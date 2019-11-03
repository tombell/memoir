package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tombell/trek"

	"github.com/tombell/memoir/pkg/config"
)

const rollbackHelpText = `usage: memoir-db rollback [<args>]

  --config  Path to .env.toml configuration file
  --steps   Number of migrations to roll back (default: 1)

Special options:

  --help    Show this message, then exit
`

func rollback(logger *log.Logger) error {
	cmd := flag.NewFlagSet("rollback", flag.ExitOnError)
	cmd.Usage = usage(rollbackHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")
	steps := cmd.Int("steps", 1, "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("cmd parse failed: %w", err)
	}

	if !cmd.Parsed() {
		return nil
	}

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	if err := trek.Rollback(logger, "postgres", cfg.DB, cfg.Migrations, *steps); err != nil {
		return fmt.Errorf("trek rollback failed: %w", err)
	}

	return nil
}
