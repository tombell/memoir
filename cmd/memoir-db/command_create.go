package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/tombell/memoir/pkg/config"
)

const createHelpText = `usage: memoir-db create [<args>]

  --config  Path to .env.toml configuration file

Special options:

  --help    Show this message, then exit
`

func create() error {
	cmd := flag.NewFlagSet("create", flag.ExitOnError)
	cmd.Usage = usage(createHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("cmd parse failed: %w", err)
	}

	if !cmd.Parsed() {
		return nil
	}

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	match := matchDBNameRegexp.FindStringSubmatch(cfg.DB)
	if match == nil {
		return errors.New("unable to find the database name from configuration file")
	}

	return exec.Command("createdb", match[1]).Run()
}
