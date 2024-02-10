package commands

import (
	"flag"
	"log/slog"
	"os"
	"os/exec"

	"github.com/tombell/memoir/internal/config"
)

const createHelpText = `usage: memoir db:create [<args>]

  --config  Path to .env.toml configuration file

Special options:

  --help    Show this message, then exit
`

func DatabaseCreateCommand(logger *slog.Logger) {
	cmd := flag.NewFlagSet("create", flag.ExitOnError)
	cmd.Usage = usageText(createHelpText)

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
		logger.Error("config load faileed", "err", err)
		return
	}

	match := matchDBNameRegexp.FindStringSubmatch(cfg.DB)
	if match == nil {
		logger.Error("unable to find the database name from configuration file")
		return
	}

	if err := exec.Command("createdb", match[1]).Run(); err != nil {
		logger.Error("unable to create database", "err", err)
	}
}
