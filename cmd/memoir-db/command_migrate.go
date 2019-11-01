package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tombell/trek"

	"github.com/tombell/memoir/config"
)

const migrateHelpText = `usage: memoir-db migrate [<args>]

  --config  Path to .env.toml configuration file

Special options:

  --help    Show this message, then exit
`

func migrate(logger *log.Logger) error {
	cmd := flag.NewFlagSet("migrate", flag.ExitOnError)
	cmd.Usage = usage(migrateHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")

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

	if err := trek.Migrate(logger, "postgres", cfg.DB, cfg.Migrations); err != nil {
		return fmt.Errorf("trek migrate failed: %w", err)
	}

	return nil
}
