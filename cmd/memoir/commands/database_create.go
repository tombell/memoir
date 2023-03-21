package commands

import (
	"flag"
	"os"
	"os/exec"

	"github.com/charmbracelet/log"

	"github.com/tombell/memoir/internal/config"
)

const createHelpText = `usage: memoir db:create [<args>]

  --config  Path to .env.toml configuration file

Special options:

  --help    Show this message, then exit
`

func DatabaseCreateCommand(logger *log.Logger) {
	cmd := flag.NewFlagSet("create", flag.ExitOnError)
	cmd.Usage = usageText(createHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		logger.Fatal("cmd parse failed", "err", err)
	}

	if !cmd.Parsed() {
		return
	}

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		logger.Fatal("config load faileed", "err", err)
	}

	match := matchDBNameRegexp.FindStringSubmatch(cfg.DB)
	if match == nil {
		logger.Fatal("unable to find the database name from configuration file")
	}

	if err := exec.Command("createdb", match[1]).Run(); err != nil {
		logger.Fatal("unable to create database", "err", err)
	}
}
