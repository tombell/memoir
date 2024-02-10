package commands

import (
	"flag"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tombell/trek"

	"github.com/tombell/memoir/internal/config"
)

const rollbackHelpText = `usage: memoir-db rollback [<args>]

  --config  Path to .env.toml configuration file
  --steps   Number of migrations to roll back (default: 1)

Special options:

  --help    Show this message, then exit
`

func DatabaseRollback(logger *slog.Logger) {
	cmd := flag.NewFlagSet("rollback", flag.ExitOnError)
	cmd.Usage = usageText(rollbackHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")
	steps := cmd.Int("steps", 1, "")

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

	if err := trek.Rollback("pgx", cfg.DB, cfg.Migrations, *steps); err != nil {
		logger.Error("trek migrate failed", "err", err)
	}
}
