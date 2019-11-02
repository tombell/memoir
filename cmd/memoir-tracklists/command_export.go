package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tombell/memoir/pkg/config"
	"github.com/tombell/memoir/pkg/datastore"
	"github.com/tombell/memoir/pkg/services"
)

const exportHelpText = `usage: memoir-tracklists export [<args>]

  --config     Path to .env.toml configuration file
  --tracklist  Name of the tracklist to export

Special options:

  --help       Show this message, then exit
`

func export() error {
	cmd := flag.NewFlagSet("export", flag.ExitOnError)
	cmd.Usage = usage(listHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")
	tracklist := cmd.String("tracklist", "", "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("cmd parse failed: %w", err)
	}

	if *tracklist == "" {
		cmd.Usage()
	}

	if !cmd.Parsed() {
		return nil
	}

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	store, err := datastore.New(cfg.DB)
	if err != nil {
		return err
	}
	defer store.Close()

	svc := services.Services{
		DataStore: store,
	}

	return svc.ExportTracklist(*tracklist, os.Stdout)
}
