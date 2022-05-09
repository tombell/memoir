package commands

import (
	"flag"
	"log"
	"os"

	"github.com/tombell/trek"

	"github.com/tombell/memoir/internal/config"
)

const rollbackHelpText = `usage: memoir-db rollback [<args>]

  --config  Path to .env.toml configuration file
  --steps   Number of migrations to roll back (default: 1)

Special options:

  --help    Show this message, then exit
`

// DatabaseRollback ...
func DatabaseRollback(logger *log.Logger) {
	cmd := flag.NewFlagSet("rollback", flag.ExitOnError)
	cmd.Usage = usageText(rollbackHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")
	steps := cmd.Int("steps", 1, "")

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

	if err := trek.Rollback("postgres", cfg.DB, cfg.Migrations, *steps); err != nil {
		logger.Fatalf("error: trek rollback failed: %s", err)
	}
}
