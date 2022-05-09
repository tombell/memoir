package commands

import (
	"flag"
	"log"
	"os"
	"os/exec"

	"github.com/tombell/memoir/internal/config"
)

const dropHelpText = `usage: memoir db:drop [<args>]

  --config  Path to .env.toml configuration file

Special options:

  --help    Show this message, then exit
`

// DatabaseDrop ...
func DatabaseDrop(logger *log.Logger) {
	cmd := flag.NewFlagSet("drop", flag.ExitOnError)
	cmd.Usage = usageText(dropHelpText)

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

	if err := exec.Command("dropdb", match[1]).Run(); err != nil {
		logger.Fatalf("error: unable to drop database: %s", err)
	}
}
