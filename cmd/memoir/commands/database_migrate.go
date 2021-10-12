package commands

import (
	"flag"
	"log"
	"os"

	"github.com/tombell/memoir/pkg/config"
	"github.com/tombell/trek"
)

const migrateHelpText = `usage: memoir-db migrate [<args>]

  --config  Path to .env.toml configuration file

Special options:

  --help    Show this message, then exit
`

// DatabaseMigrate ...
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
