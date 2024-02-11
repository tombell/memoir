package commands

import (
	"flag"
	"log/slog"
	"os"

	"github.com/tombell/trek"

	"github.com/tombell/memoir/internal/config"
)

func DatabaseMigrate(logger *slog.Logger) {
	cmd := flag.NewFlagSet("migrate", flag.ExitOnError)

	cfgpath := cmd.String("config", "config.dev.json", "")

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
