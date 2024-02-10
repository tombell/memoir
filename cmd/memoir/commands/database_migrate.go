package commands

import (
	"flag"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tombell/trek"

	"github.com/tombell/memoir/internal/config"
)

const migrateHelpText = `usage: memoir-db migrate [<args>]

  --config  Path to .env.toml configuration file

Special options:

  --help    Show this message, then exit
`

func DatabaseMigrate(logger *slog.Logger) {
	cmd := flag.NewFlagSet("migrate", flag.ExitOnError)
	cmd.Usage = usageText(migrateHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		logger.Error("cmd parse failed", "err", err)
		return
	}

	if !cmd.Parsed() {
		return
	}

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		logger.Error("config load failed", "err", err)
		return
	}

	if err := trek.Migrate("pgx", cfg.DB, cfg.Migrations); err != nil {
		logger.Error("trek migrate failed", "err", err)
	}
}
