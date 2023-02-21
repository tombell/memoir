package commands

import (
	"flag"
	"os"

	"github.com/charmbracelet/log"
	"github.com/tombell/trek"

	"github.com/tombell/memoir/internal/config"
)

const migrateHelpText = `usage: memoir-db migrate [<args>]

  --config  Path to .env.toml configuration file

Special options:

  --help    Show this message, then exit
`

func DatabaseMigrate(logger log.Logger) {
	cmd := flag.NewFlagSet("migrate", flag.ExitOnError)
	cmd.Usage = usageText(migrateHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		logger.Fatal("cmd parse failed", "err", err)
	}

	if !cmd.Parsed() {
		return
	}

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		logger.Fatal("config load failed", "err", err)
	}

	if err := trek.Migrate("postgres", cfg.DB, cfg.Migrations); err != nil {
		logger.Fatal("trek migrate failed", "err", err)
	}
}
