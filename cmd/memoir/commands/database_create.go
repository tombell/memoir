package commands

import (
	"flag"
	"log"
	"os"
	"os/exec"

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
		logger.Fatalf("error: cmd parse failed: %s", err)
	}

	if !cmd.Parsed() {
		return
	}

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		logger.Fatalf("error: config load failed: %s", err)
	}

	match := matchDBNameRegexp.FindStringSubmatch(cfg.DB)
	if match == nil {
		logger.Fatalf("error: unable to find the database name from configuration file")
	}

	if err := exec.Command("createdb", match[1]).Run(); err != nil {
		logger.Fatalf("error: unable to create database: %s", err)
	}
}
