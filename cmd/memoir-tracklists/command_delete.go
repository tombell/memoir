package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tombell/memoir/pkg/config"
	"github.com/tombell/memoir/pkg/datastore"
	"github.com/tombell/memoir/pkg/services"
)

const deleteHelpText = `usage: memoir-tracklists delete [<args>]

  --config     Path to .env.toml configuration file
  --tracklist  Name of the tracklist to delete

Special options:

  --help       Show this message, then exit
`

func delete() error {
	cmd := flag.NewFlagSet("delete", flag.ExitOnError)
	cmd.Usage = usage(deleteHelpText)

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
	defer store.Close()
	if err != nil {
		return err
	}

	svc := services.Services{
		DataStore: store,
	}

	return svc.RemoveTracklist(*tracklist)
}
