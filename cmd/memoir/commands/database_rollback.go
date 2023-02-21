package commands

import (
	"flag"
	"os"

	"github.com/charmbracelet/log"
	"github.com/tombell/trek"

	"github.com/tombell/memoir/internal/config"
)

const rollbackHelpText = `usage: memoir-db rollback [<args>]

  --config  Path to .env.toml configuration file
  --steps   Number of migrations to roll back (default: 1)

Special options:

  --help    Show this message, then exit
`

func DatabaseRollback(logger log.Logger) {
	cmd := flag.NewFlagSet("rollback", flag.ExitOnError)
	cmd.Usage = usageText(rollbackHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")
	steps := cmd.Int("steps", 1, "")

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

	if err := trek.Rollback("postgres", cfg.DB, cfg.Migrations, *steps); err != nil {
		logger.Fatal("trek migrate failed", "err", err)
	}
}
