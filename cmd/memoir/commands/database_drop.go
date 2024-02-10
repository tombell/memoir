package commands

import (
	"flag"
	"log/slog"
	"os"
	"os/exec"

	"github.com/tombell/memoir/internal/config"
)

const dropHelpText = `usage: memoir db:drop [<args>]

  --config  Path to .env.toml configuration file

Special options:

  --help    Show this message, then exit
`

func DatabaseDrop(logger *slog.Logger) {
	cmd := flag.NewFlagSet("drop", flag.ExitOnError)
	cmd.Usage = usageText(dropHelpText)

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

	match := matchDBNameRegexp.FindStringSubmatch(cfg.DB)
	if match == nil {
		logger.Error("unable to find the database name from configuration file")
		return
	}

	if err := exec.Command("dropdb", match[1]).Run(); err != nil {
		logger.Error("error: unable to drop database", "err", err)
	}
}
