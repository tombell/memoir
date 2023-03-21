package commands

import (
	"flag"
	"os"
	"os/exec"

	"github.com/charmbracelet/log"

	"github.com/tombell/memoir/internal/config"
)

const dropHelpText = `usage: memoir db:drop [<args>]

  --config  Path to .env.toml configuration file

Special options:

  --help    Show this message, then exit
`

func DatabaseDrop(logger *log.Logger) {
	cmd := flag.NewFlagSet("drop", flag.ExitOnError)
	cmd.Usage = usageText(dropHelpText)

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

	match := matchDBNameRegexp.FindStringSubmatch(cfg.DB)
	if match == nil {
		logger.Fatal("unable to find the database name from configuration file")
	}

	if err := exec.Command("dropdb", match[1]).Run(); err != nil {
		logger.Fatal("error: unable to drop database", "err", err)
	}
}
