package commands

import (
	"flag"
	"log"
	"os"

	"github.com/tombell/trek"

	"github.com/tombell/memoir/internal/config"
)

const migrateHelpText = `usage: memoir-db migrate [<args>]

  --config  Path to .env.toml configuration file

Special options:

  --help    Show this message, then exit
`

func DatabaseMigrate(logger *log.Logger) {
	cmd := flag.NewFlagSet("migrate", flag.ExitOnError)
	cmd.Usage = usageText(migrateHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		logger.Fatalf("error: cmd parse failed: %s", err)
	}

	if !cmd.Parsed() {
		return
	}

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		logger.Fatalf("error: config load failed: %s", err)
	}

	if err := trek.Migrate("postgres", cfg.DB, cfg.Migrations); err != nil {
		logger.Fatalf("error: trek migrate failed: %s", err)
	}
}
